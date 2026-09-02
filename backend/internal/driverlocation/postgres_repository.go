package driverlocation

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) UpsertCurrent(ctx context.Context, driverUserID uuid.UUID, input Input) (Location, error) {
	var location Location
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO driver_locations (driver_user_id, latitude, longitude, updated_at)
		SELECT dp.user_id, $2, $3, NOW()
		FROM driver_profiles dp
		WHERE dp.user_id = $1
		  AND dp.status = 'active'
		ON CONFLICT (driver_user_id) DO UPDATE SET
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			updated_at = NOW()
		RETURNING driver_user_id, latitude, longitude, updated_at
	`, driverUserID, input.Latitude, input.Longitude).Scan(
		&location.DriverUserID,
		&location.Latitude,
		&location.Longitude,
		&location.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Location{}, ErrDriverNotFound
	}
	if err != nil {
		return Location{}, err
	}
	return location, nil
}
