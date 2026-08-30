package ride

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) Create(ctx context.Context, riderUserID uuid.UUID, pickup, destination Location) (Request, error) {
	var request Request
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ride_requests (
			id, rider_user_id,
			pickup_latitude, pickup_longitude,
			destination_latitude, destination_longitude,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, rider_user_id,
		          pickup_latitude, pickup_longitude,
		          destination_latitude, destination_longitude,
		          status, created_at
	`, uuid.New(), riderUserID,
		pickup.Latitude, pickup.Longitude,
		destination.Latitude, destination.Longitude,
		StatusRequested,
	).Scan(
		&request.ID, &request.RiderUserID,
		&request.Pickup.Latitude, &request.Pickup.Longitude,
		&request.Destination.Latitude, &request.Destination.Longitude,
		&request.Status, &request.CreatedAt,
	)
	if err != nil {
		return Request{}, err
	}
	return request, nil
}
