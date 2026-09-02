package riderlocation

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) GetForActiveTrip(ctx context.Context, riderUserID uuid.UUID) (View, error) {
	var view View
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT
			t.ride_request_id,
			dl.latitude,
			dl.longitude,
			dl.updated_at
		FROM trips t
		LEFT JOIN driver_locations dl ON dl.driver_user_id = t.driver_user_id
		WHERE t.rider_user_id = $1
		  AND t.status IN ('assigned', 'in_progress')
		LIMIT 1
	`, riderUserID).Scan(
		&view.RideRequestID,
		&latitude,
		&longitude,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return View{}, ErrActiveTripNotFound
	}
	if err != nil {
		return View{}, err
	}
	if !latitude.Valid || !longitude.Valid || !updatedAt.Valid {
		return View{}, ErrLocationNotFound
	}

	view.Latitude = latitude.Float64
	view.Longitude = longitude.Float64
	view.UpdatedAt = updatedAt.Time
	return view, nil
}
