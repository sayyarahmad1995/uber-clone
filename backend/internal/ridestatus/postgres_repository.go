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

type scanner interface { Scan(dest ...any) error }

func scanView(row scanner) (View, error) {
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
	if err := row.Scan(
		&view.RideRequest.ID,
		&view.RideRequest.RiderUserID,
		&view.RideRequest.Pickup.Latitude,
		&view.RideRequest.Pickup.Longitude,
		&view.RideRequest.Destination.Latitude,
		&view.RideRequest.Destination.Longitude,
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
	); err != nil { return View{}, err }
	if proposedAmount.Valid && proposedCurrency.Valid {
		view.RideRequest.ProposedFare = &ride.Money{AmountMinor: proposedAmount.Int64, Currency: proposedCurrency.String}
	}
	if rideCancelledAt.Valid { view.RideRequest.CancelledAt = &rideCancelledAt.Time }
	if rideCancelledBy.Valid { view.RideRequest.CancelledBy = ride.CancellationActor(rideCancelledBy.String) }
	if tripDriver.Valid && tripStatus.Valid && assignedAt.Valid {
		projectedTrip := trip.Trip{RideRequestID: view.RideRequest.ID, RiderUserID: view.RideRequest.RiderUserID, DriverUserID: tripDriver.UUID, Status: trip.Status(tripStatus.String), AssignedAt: assignedAt.Time}
		if startedAt.Valid { projectedTrip.StartedAt = &startedAt.Time }
		if completedAt.Valid { projectedTrip.CompletedAt = &completedAt.Time }
		if tripCancelledAt.Valid { projectedTrip.CancelledAt = &tripCancelledAt.Time }
		view.Trip = &projectedTrip
	}
	return view, nil
}

const viewColumns = `
	rr.id,
	rr.rider_user_id,
	rr.pickup_latitude,
	rr.pickup_longitude,
	rr.destination_latitude,
	rr.destination_longitude,
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
	t.cancelled_at`

func (r PostgresRepository) GetOwned(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (View, error) {
	view, err := scanView(r.db.QueryRowContext(ctx, `SELECT `+viewColumns+` FROM ride_requests rr LEFT JOIN trips t ON t.ride_request_id = rr.id WHERE rr.id = $1 AND rr.rider_user_id = $2`, rideRequestID, riderUserID))
	if errors.Is(err, sql.ErrNoRows) { return View{}, ErrNotFound }
	return view, err
}

func (r PostgresRepository) ListOwned(ctx context.Context, riderUserID uuid.UUID, limit int) ([]View, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+viewColumns+` FROM ride_requests rr LEFT JOIN trips t ON t.ride_request_id = rr.id WHERE rr.rider_user_id = $1 ORDER BY rr.created_at DESC, rr.id DESC LIMIT $2`, riderUserID, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	views := make([]View, 0)
	for rows.Next() {
		view, err := scanView(rows)
		if err != nil { return nil, err }
		views = append(views, view)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return views, nil
}
