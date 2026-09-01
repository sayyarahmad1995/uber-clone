package drivertrip

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) GetCurrent(ctx context.Context, driverUserID uuid.UUID) (View, error) {
	var view View
	err := r.db.QueryRowContext(ctx, `
		SELECT
			t.ride_request_id,
			rr.pickup_latitude,
			rr.pickup_longitude,
			rr.destination_latitude,
			rr.destination_longitude,
			t.status,
			t.assigned_at,
			t.started_at
		FROM trips t
		JOIN ride_requests rr ON rr.id = t.ride_request_id
		WHERE t.driver_user_id = $1
		  AND t.status IN ('assigned', 'in_progress')
	`, driverUserID).Scan(
		&view.RideRequestID,
		&view.Pickup.Latitude,
		&view.Pickup.Longitude,
		&view.Destination.Latitude,
		&view.Destination.Longitude,
		&view.Status,
		&view.AssignedAt,
		&view.StartedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return View{}, ErrNotFound
	}
	if err != nil {
		return View{}, err
	}
	return view, nil
}
