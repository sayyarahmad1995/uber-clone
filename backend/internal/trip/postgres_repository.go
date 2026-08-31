package trip

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) Accept(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Acceptance, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Acceptance{}, err
	}
	defer tx.Rollback()

	var candidateStatus string
	var candidateCreatedAt time.Time
	var candidateDecidedAt *time.Time
	if err := tx.QueryRowContext(ctx, `
		SELECT status, created_at, decided_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(&candidateStatus, &candidateCreatedAt, &candidateDecidedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Acceptance{}, ErrAssignmentNotFound
		}
		return Acceptance{}, err
	}

	switch candidateStatus {
	case "pending":
		if err := tx.QueryRowContext(ctx, `
			UPDATE ride_driver_candidates
			SET status = 'accepted', decided_at = NOW()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING decided_at
		`, rideRequestID, driverUserID).Scan(&candidateDecidedAt); err != nil {
			return Acceptance{}, err
		}
	case "accepted":
		// Idempotent acceptance: preserve the original decision timestamp.
	default:
		return Acceptance{}, ErrAssignmentResolved
	}

	var riderUserID uuid.UUID
	if err := tx.QueryRowContext(ctx, `
		SELECT rider_user_id
		FROM ride_requests
		WHERE id = $1
	`, rideRequestID).Scan(&riderUserID); err != nil {
		return Acceptance{}, err
	}

	assignedAt := candidateCreatedAt
	if candidateDecidedAt != nil {
		assignedAt = *candidateDecidedAt
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO trips (ride_request_id, rider_user_id, driver_user_id, assigned_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (ride_request_id) DO NOTHING
	`, rideRequestID, riderUserID, driverUserID, assignedAt); err != nil {
		return Acceptance{}, err
	}

	trip, err := selectTrip(ctx, tx, rideRequestID, driverUserID, false)
	if err != nil {
		return Acceptance{}, err
	}
	if err := tx.Commit(); err != nil {
		return Acceptance{}, err
	}
	return Acceptance{
		Trip:               trip,
		CandidateCreatedAt: candidateCreatedAt,
		CandidateDecidedAt: candidateDecidedAt,
	}, nil
}

func (r PostgresRepository) Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()

	trip, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}

	switch trip.Status {
	case StatusAssigned:
		if err := tx.QueryRowContext(ctx, `
			UPDATE trips
			SET status = 'in_progress', started_at = NOW()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at
		`, rideRequestID, driverUserID).Scan(
			&trip.RideRequestID,
			&trip.RiderUserID,
			&trip.DriverUserID,
			&trip.Status,
			&trip.AssignedAt,
			&trip.StartedAt,
			&trip.CompletedAt,
		); err != nil {
			return Trip{}, err
		}
	case StatusInProgress:
		// Idempotent start.
	case StatusCompleted:
		return Trip{}, ErrTripCompleted
	default:
		return Trip{}, errors.New("unknown trip status")
	}

	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return trip, nil
}

func (r PostgresRepository) Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()

	trip, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}

	switch trip.Status {
	case StatusAssigned:
		return Trip{}, ErrTripNotStarted
	case StatusInProgress:
		if err := tx.QueryRowContext(ctx, `
			UPDATE trips
			SET status = 'completed', completed_at = NOW()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at
		`, rideRequestID, driverUserID).Scan(
			&trip.RideRequestID,
			&trip.RiderUserID,
			&trip.DriverUserID,
			&trip.Status,
			&trip.AssignedAt,
			&trip.StartedAt,
			&trip.CompletedAt,
		); err != nil {
			return Trip{}, err
		}
	case StatusCompleted:
		// Idempotent completion; also repair a missing release marker below.
	default:
		return Trip{}, errors.New("unknown trip status")
	}

	if trip.CompletedAt == nil {
		return Trip{}, errors.New("completed trip missing completed_at")
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_candidates
		SET released_at = COALESCE(released_at, $3)
		WHERE ride_request_id = $1
		  AND driver_user_id = $2
		  AND status = 'accepted'
	`, rideRequestID, driverUserID, *trip.CompletedAt); err != nil {
		return Trip{}, err
	}

	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return trip, nil
}

func selectTrip(ctx context.Context, tx *sql.Tx, rideRequestID, driverUserID uuid.UUID, forUpdate bool) (Trip, error) {
	query := `
		SELECT ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at
		FROM trips
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	var trip Trip
	if err := tx.QueryRowContext(ctx, query, rideRequestID, driverUserID).Scan(
		&trip.RideRequestID,
		&trip.RiderUserID,
		&trip.DriverUserID,
		&trip.Status,
		&trip.AssignedAt,
		&trip.StartedAt,
		&trip.CompletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Trip{}, ErrTripNotFound
		}
		return Trip{}, err
	}
	return trip, nil
}
