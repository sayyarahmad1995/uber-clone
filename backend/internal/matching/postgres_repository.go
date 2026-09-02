package matching

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PostgresRepository struct {
	db                      *sql.DB
	driverLocationFreshness time.Duration
}

func NewPostgresRepository(db *sql.DB, driverLocationFreshness time.Duration) PostgresRepository {
	return PostgresRepository{db: db, driverLocationFreshness: driverLocationFreshness}
}

func (r PostgresRepository) Match(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()

	var ownedRideID uuid.UUID
	var bookingMode, rideStatus string
	var pickupLatitude, pickupLongitude float64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id, booking_mode, status, pickup_latitude, pickup_longitude FROM ride_requests WHERE id = $1 AND rider_user_id = $2 FOR UPDATE`,
		rideRequestID,
		riderUserID,
	).Scan(&ownedRideID, &bookingMode, &rideStatus, &pickupLatitude, &pickupLongitude); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, ErrRideNotFound
		}
		return Result{}, err
	}
	if bookingMode != "automatic" {
		return Result{}, ErrRideNotMatchable
	}
	if rideStatus != "requested" {
		return Result{}, ErrRideNotOpen
	}

	var existing Candidate
	err = tx.QueryRowContext(ctx, `
		SELECT ride_request_id, driver_user_id, status, created_at, decided_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1
		  AND status IN ('pending', 'accepted')
		  AND released_at IS NULL
		ORDER BY created_at DESC, driver_user_id ASC
		LIMIT 1
	`, rideRequestID).Scan(
		&existing.RideRequestID,
		&existing.DriverUserID,
		&existing.Status,
		&existing.CreatedAt,
		&existing.DecidedAt,
	)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return Result{}, err
		}
		return Result{Candidate: existing, Created: false}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	locationCutoff := time.Now().UTC().Add(-r.driverLocationFreshness)
	var driverUserID uuid.UUID
	if err := tx.QueryRowContext(ctx, `
		SELECT p.user_id
		FROM driver_profiles p
		JOIN driver_vehicles v ON v.driver_user_id = p.user_id
		JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
		JOIN driver_locations dl ON dl.driver_user_id = p.user_id
		WHERE p.status = 'active'
		  AND p.is_online = TRUE
		  AND p.user_id <> $1
		  AND dl.updated_at >= $3
		  AND NOT EXISTS (
			SELECT 1
			FROM ride_driver_candidates prior
			WHERE prior.ride_request_id = $2
			  AND prior.driver_user_id = p.user_id
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM ride_driver_candidates active
			WHERE active.driver_user_id = p.user_id
			  AND active.status IN ('pending', 'accepted')
			  AND active.released_at IS NULL
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM trips t
			WHERE t.driver_user_id = p.user_id
			  AND t.status IN ('assigned', 'in_progress')
		  )
		ORDER BY
		  2 * 6371000 * ASIN(SQRT(
			POWER(SIN(RADIANS(dl.latitude - $4) / 2), 2) +
			COS(RADIANS($4)) * COS(RADIANS(dl.latitude)) *
			POWER(SIN(RADIANS(dl.longitude - $5) / 2), 2)
		  )) ASC,
		  p.user_id ASC
		LIMIT 1
		FOR UPDATE OF p SKIP LOCKED
	`, riderUserID, rideRequestID, locationCutoff, pickupLatitude, pickupLongitude).Scan(&driverUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Result{}, ErrNoEligibleDriver
		}
		return Result{}, err
	}

	var candidate Candidate
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ride_driver_candidates (ride_request_id, driver_user_id)
		VALUES ($1, $2)
		RETURNING ride_request_id, driver_user_id, status, created_at, decided_at
	`, rideRequestID, driverUserID).Scan(
		&candidate.RideRequestID,
		&candidate.DriverUserID,
		&candidate.Status,
		&candidate.CreatedAt,
		&candidate.DecidedAt,
	); err != nil {
		return Result{}, err
	}

	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{Candidate: candidate, Created: true}, nil
}

func (r PostgresRepository) Reject(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Candidate, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Candidate{}, err
	}
	defer tx.Rollback()

	var candidate Candidate
	var releasedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `
		SELECT ride_request_id, driver_user_id, status, created_at, decided_at, released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(
		&candidate.RideRequestID,
		&candidate.DriverUserID,
		&candidate.Status,
		&candidate.CreatedAt,
		&candidate.DecidedAt,
		&releasedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Candidate{}, ErrCandidateNotFound
		}
		return Candidate{}, err
	}
	if releasedAt.Valid {
		return Candidate{}, ErrCandidateResolved
	}

	if candidate.Status == CandidateStatusRejected {
		if err := tx.Commit(); err != nil {
			return Candidate{}, err
		}
		return candidate, nil
	}
	if candidate.Status != CandidateStatusPending {
		return Candidate{}, ErrCandidateResolved
	}

	if err := tx.QueryRowContext(ctx, `
		UPDATE ride_driver_candidates
		SET status = 'rejected', decided_at = NOW()
		WHERE ride_request_id = $1 AND driver_user_id = $2
		RETURNING ride_request_id, driver_user_id, status, created_at, decided_at
	`, rideRequestID, driverUserID).Scan(
		&candidate.RideRequestID,
		&candidate.DriverUserID,
		&candidate.Status,
		&candidate.CreatedAt,
		&candidate.DecidedAt,
	); err != nil {
		return Candidate{}, err
	}

	if err := tx.Commit(); err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}
