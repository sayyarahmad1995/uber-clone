package trip

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PostgresRepository struct {
	db                       *sql.DB
	candidateResponseTimeout time.Duration
}

func NewPostgresRepository(db *sql.DB, candidateResponseTimeout time.Duration) PostgresRepository {
	return PostgresRepository{db: db, candidateResponseTimeout: candidateResponseTimeout}
}

func (r PostgresRepository) Accept(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Acceptance, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Acceptance{}, err
	}
	defer tx.Rollback()

	var riderUserID uuid.UUID
	var rideStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT rider_user_id, status
		FROM ride_requests
		WHERE id = $1
		FOR UPDATE
	`, rideRequestID).Scan(&riderUserID, &rideStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Acceptance{}, ErrAssignmentNotFound
		}
		return Acceptance{}, err
	}
	if rideStatus != "requested" {
		return Acceptance{}, ErrAssignmentResolved
	}

	var candidateStatus string
	var candidateCreatedAt time.Time
	var candidateDecidedAt *time.Time
	var candidateReleasedAt *time.Time
	if err := tx.QueryRowContext(ctx, `
		SELECT status, created_at, decided_at, released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(&candidateStatus, &candidateCreatedAt, &candidateDecidedAt, &candidateReleasedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Acceptance{}, ErrAssignmentNotFound
		}
		return Acceptance{}, err
	}
	if candidateReleasedAt != nil {
		return Acceptance{}, ErrAssignmentResolved
	}

	switch candidateStatus {
	case "pending":
		now := time.Now().UTC()
		if !candidateCreatedAt.After(now.Add(-r.candidateResponseTimeout)) {
			if _, err := tx.ExecContext(ctx, `
				UPDATE ride_driver_candidates
				SET released_at = $3
				WHERE ride_request_id = $1
				  AND driver_user_id = $2
				  AND status = 'pending'
				  AND released_at IS NULL
			`, rideRequestID, driverUserID, now); err != nil {
				return Acceptance{}, err
			}
			if err := tx.Commit(); err != nil {
				return Acceptance{}, err
			}
			return Acceptance{}, ErrAssignmentResolved
		}
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

	assignedTrip, err := selectTrip(ctx, tx, rideRequestID, driverUserID, false)
	if err != nil {
		return Acceptance{}, err
	}
	if assignedTrip.Status == StatusCancelled {
		return Acceptance{}, ErrAssignmentResolved
	}
	if err := tx.Commit(); err != nil {
		return Acceptance{}, err
	}
	return Acceptance{
		Trip:               assignedTrip,
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
		// Idempotent completion; also repair a missing release marker below.
	case StatusCancelled:
		return Trip{}, ErrTripCancelled
	default:
		return Trip{}, errors.New("unknown trip status")
	}

	if result.CompletedAt == nil {
		return Trip{}, errors.New("completed trip missing completed_at")
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_candidates
		SET released_at = COALESCE(released_at, $3)
		WHERE ride_request_id = $1
		  AND driver_user_id = $2
		  AND status = 'accepted'
	`, rideRequestID, driverUserID, *result.CompletedAt); err != nil {
		return Trip{}, err
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
