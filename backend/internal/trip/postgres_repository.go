package trip

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return PostgresRepository{db: db}
}

func (r PostgresRepository) Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()

	result, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}

	switch result.Status {
	case StatusAssigned:
		if err := tx.QueryRowContext(ctx, `
			UPDATE trips
			SET status = 'in_progress', started_at = NOW()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at, cancelled_at
		`, rideRequestID, driverUserID).Scan(
			&result.RideRequestID,
			&result.RiderUserID,
			&result.DriverUserID,
			&result.Status,
			&result.AssignedAt,
			&result.StartedAt,
			&result.CompletedAt,
			&result.CancelledAt,
		); err != nil {
			return Trip{}, err
		}
	case StatusInProgress:
		// Idempotent start.
	case StatusCompleted:
		return Trip{}, ErrTripCompleted
	case StatusCancelled:
		return Trip{}, ErrTripCancelled
	default:
		return Trip{}, errors.New("unknown trip status")
	}

	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return result, nil
}

func (r PostgresRepository) Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()

	result, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}

	switch result.Status {
	case StatusAssigned:
		return Trip{}, ErrTripNotStarted
	case StatusInProgress:
		if err := tx.QueryRowContext(ctx, `
			UPDATE trips
			SET status = 'completed', completed_at = NOW()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at, cancelled_at
		`, rideRequestID, driverUserID).Scan(
			&result.RideRequestID,
			&result.RiderUserID,
			&result.DriverUserID,
			&result.Status,
			&result.AssignedAt,
			&result.StartedAt,
			&result.CompletedAt,
			&result.CancelledAt,
		); err != nil {
			return Trip{}, err
		}
	case StatusCompleted:
		// Idempotent completion.
	case StatusCancelled:
		return Trip{}, ErrTripCancelled
	default:
		return Trip{}, errors.New("unknown trip status")
	}

	if result.CompletedAt == nil {
		return Trip{}, errors.New("completed trip missing completed_at")
	}

	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return result, nil
}

func selectTrip(ctx context.Context, tx *sql.Tx, rideRequestID, driverUserID uuid.UUID, forUpdate bool) (Trip, error) {
	query := `
		SELECT ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at, cancelled_at
		FROM trips
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	var result Trip
	if err := tx.QueryRowContext(ctx, query, rideRequestID, driverUserID).Scan(
		&result.RideRequestID,
		&result.RiderUserID,
		&result.DriverUserID,
		&result.Status,
		&result.AssignedAt,
		&result.StartedAt,
		&result.CompletedAt,
		&result.CancelledAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Trip{}, ErrTripNotFound
		}
		return Trip{}, err
	}
	return result, nil
}
