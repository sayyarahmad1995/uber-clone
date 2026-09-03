package cancellation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

type lockedRide struct {
	status      ride.Status
	cancelledBy sql.NullString
	cancelledAt sql.NullTime
}

func (r PostgresRepository) CancelByRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()

	state, err := lockRide(ctx, tx, rideRequestID, &riderUserID)
	if err != nil {
		return Result{}, err
	}
	assignedTrip, found, err := lockTrip(ctx, tx, rideRequestID, nil)
	if err != nil {
		return Result{}, err
	}

	if state.status == ride.StatusCancelled {
		result, err := existingResult(rideRequestID, state, assignedTrip, found)
		if err != nil {
			return Result{}, err
		}
		if err := tx.Commit(); err != nil {
			return Result{}, err
		}
		return result, nil
	}
	if state.status != ride.StatusRequested {
		return Result{}, fmt.Errorf("unknown ride request status %q", state.status)
	}
	if found && assignedTrip.Status == trip.StatusCompleted {
		return Result{}, ErrTripCompleted
	}

	result, err := cancelLocked(ctx, tx, rideRequestID, ride.CancellationActorRider, assignedTrip, found)
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (r PostgresRepository) CancelByDriver(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Result, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()

	state, err := lockRide(ctx, tx, rideRequestID, nil)
	if err != nil {
		return Result{}, err
	}
	assignedTrip, found, err := lockTrip(ctx, tx, rideRequestID, &driverUserID)
	if err != nil {
		return Result{}, err
	}
	if !found {
		return Result{}, ErrNotFound
	}

	if state.status == ride.StatusCancelled {
		result, err := existingResult(rideRequestID, state, assignedTrip, true)
		if err != nil {
			return Result{}, err
		}
		if err := tx.Commit(); err != nil {
			return Result{}, err
		}
		return result, nil
	}
	if state.status != ride.StatusRequested {
		return Result{}, fmt.Errorf("unknown ride request status %q", state.status)
	}
	if assignedTrip.Status == trip.StatusCompleted {
		return Result{}, ErrTripCompleted
	}

	result, err := cancelLocked(ctx, tx, rideRequestID, ride.CancellationActorDriver, assignedTrip, true)
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func lockRide(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID, riderUserID *uuid.UUID) (lockedRide, error) {
	query := `
		SELECT status, cancelled_by, cancelled_at
		FROM ride_requests
		WHERE id = $1
	`
	args := []any{rideRequestID}
	if riderUserID != nil {
		query += " AND rider_user_id = $2"
		args = append(args, *riderUserID)
	}
	query += " FOR UPDATE"

	var state lockedRide
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&state.status, &state.cancelledBy, &state.cancelledAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return lockedRide{}, ErrNotFound
		}
		return lockedRide{}, err
	}
	return state, nil
}

func lockTrip(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID, driverUserID *uuid.UUID) (trip.Trip, bool, error) {
	query := `
		SELECT ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at, cancelled_at
		FROM trips
		WHERE ride_request_id = $1
	`
	args := []any{rideRequestID}
	if driverUserID != nil {
		query += " AND driver_user_id = $2"
		args = append(args, *driverUserID)
	}
	query += " FOR UPDATE"

	var result trip.Trip
	if err := tx.QueryRowContext(ctx, query, args...).Scan(
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
			return trip.Trip{}, false, nil
		}
		return trip.Trip{}, false, err
	}
	return result, true, nil
}

func existingResult(rideRequestID uuid.UUID, state lockedRide, assignedTrip trip.Trip, hasTrip bool) (Result, error) {
	if !state.cancelledBy.Valid || !state.cancelledAt.Valid {
		return Result{}, errors.New("cancelled ride request missing cancellation metadata")
	}
	result := Result{
		RideRequestID: rideRequestID,
		Status:        ride.StatusCancelled,
		CancelledBy:   ride.CancellationActor(state.cancelledBy.String),
		CancelledAt:   state.cancelledAt.Time,
	}
	if hasTrip {
		if assignedTrip.Status != trip.StatusCancelled || assignedTrip.CancelledAt == nil {
			return Result{}, errors.New("cancelled ride request has non-cancelled trip")
		}
		result.Trip = &assignedTrip
	}
	return result, nil
}

func cancelLocked(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID, actor ride.CancellationActor, assignedTrip trip.Trip, hasTrip bool) (Result, error) {
	var cancelledAt time.Time
	if err := tx.QueryRowContext(ctx, `
		UPDATE ride_requests
		SET status = 'cancelled', cancelled_by = $2, cancelled_at = NOW()
		WHERE id = $1
		RETURNING cancelled_at
	`, rideRequestID, actor).Scan(&cancelledAt); err != nil {
		return Result{}, err
	}

	result := Result{
		RideRequestID: rideRequestID,
		Status:        ride.StatusCancelled,
		CancelledBy:   actor,
		CancelledAt:   cancelledAt,
	}

	if hasTrip {
		switch assignedTrip.Status {
		case trip.StatusAssigned, trip.StatusInProgress:
			if err := tx.QueryRowContext(ctx, `
				UPDATE trips
				SET status = 'cancelled', cancelled_at = $2
				WHERE ride_request_id = $1
				RETURNING ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at, cancelled_at
			`, rideRequestID, cancelledAt).Scan(
				&assignedTrip.RideRequestID,
				&assignedTrip.RiderUserID,
				&assignedTrip.DriverUserID,
				&assignedTrip.Status,
				&assignedTrip.AssignedAt,
				&assignedTrip.StartedAt,
				&assignedTrip.CompletedAt,
				&assignedTrip.CancelledAt,
			); err != nil {
				return Result{}, err
			}
		case trip.StatusCompleted:
			return Result{}, ErrTripCompleted
		case trip.StatusCancelled:
			return Result{}, errors.New("trip cancelled before ride request")
		default:
			return Result{}, fmt.Errorf("unknown trip status %q", assignedTrip.Status)
		}
		result.Trip = &assignedTrip
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE ride_offers
		SET status = 'closed', decided_at = $2, updated_at = NOW()
		WHERE ride_request_id = $1
		  AND status = 'pending'
	`, rideRequestID, cancelledAt); err != nil {
		return Result{}, err
	}

	return result, nil
}
