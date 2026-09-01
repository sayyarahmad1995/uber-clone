package ridestatus

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) GetOwned(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (View, error) {
	var view View
	var proposedAmount sql.NullInt64
	var proposedCurrency sql.NullString
	var rideCancelledAt sql.NullTime
	var rideCancelledBy sql.NullString
	var tripDriver uuid.NullUUID
	var tripStatus sql.NullString
	var assignedAt sql.NullTime
	var startedAt sql.NullTime
	var completedAt sql.NullTime
	var tripCancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
			rr.id,
			rr.rider_user_id,
			rr.pickup_latitude,
			rr.pickup_longitude,
			rr.destination_latitude,
			rr.destination_longitude,
			rr.booking_mode,
			rr.proposed_fare_minor,
			rr.currency,
			rr.status,
			rr.created_at,
			rr.cancelled_at,
			rr.cancelled_by,
			t.driver_user_id,
			t.status,
			t.assigned_at,
			t.started_at,
			t.completed_at,
			t.cancelled_at
		FROM ride_requests rr
		LEFT JOIN trips t ON t.ride_request_id = rr.id
		WHERE rr.id = $1
		  AND rr.rider_user_id = $2
	`, rideRequestID, riderUserID).Scan(
		&view.RideRequest.ID,
		&view.RideRequest.RiderUserID,
		&view.RideRequest.Pickup.Latitude,
		&view.RideRequest.Pickup.Longitude,
		&view.RideRequest.Destination.Latitude,
		&view.RideRequest.Destination.Longitude,
		&view.RideRequest.BookingMode,
		&proposedAmount,
		&proposedCurrency,
		&view.RideRequest.Status,
		&view.RideRequest.CreatedAt,
		&rideCancelledAt,
		&rideCancelledBy,
		&tripDriver,
		&tripStatus,
		&assignedAt,
		&startedAt,
		&completedAt,
		&tripCancelledAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return View{}, ErrNotFound
	}
	if err != nil {
		return View{}, err
	}

	if proposedAmount.Valid && proposedCurrency.Valid {
		view.RideRequest.ProposedFare = &ride.Money{
			AmountMinor: proposedAmount.Int64,
			Currency:    proposedCurrency.String,
		}
	}
	if rideCancelledAt.Valid {
		view.RideRequest.CancelledAt = &rideCancelledAt.Time
	}
	if rideCancelledBy.Valid {
		view.RideRequest.CancelledBy = ride.CancellationActor(rideCancelledBy.String)
	}

	if tripDriver.Valid && tripStatus.Valid && assignedAt.Valid {
		projectedTrip := trip.Trip{
			RideRequestID: view.RideRequest.ID,
			RiderUserID:   view.RideRequest.RiderUserID,
			DriverUserID:  tripDriver.UUID,
			Status:        trip.Status(tripStatus.String),
			AssignedAt:    assignedAt.Time,
		}
		if startedAt.Valid {
			projectedTrip.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			projectedTrip.CompletedAt = &completedAt.Time
		}
		if tripCancelledAt.Valid {
			projectedTrip.CancelledAt = &tripCancelledAt.Time
		}
		view.Trip = &projectedTrip
	}

	return view, nil
}
