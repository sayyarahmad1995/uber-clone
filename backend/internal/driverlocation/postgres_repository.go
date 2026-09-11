package driverlocation

import (
	"context"
	"database/sql"
	"errors"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/driver"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) UpsertCurrent(ctx context.Context, driverUserID uuid.UUID, input Input) (Location, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil { return Location{}, err }
    defer tx.Rollback()
    var id uuid.UUID
    err = tx.QueryRowContext(ctx, `SELECT user_id FROM driver_profiles WHERE user_id=$1 FOR UPDATE`,driverUserID).Scan(&id)
    if errors.Is(err,sql.ErrNoRows) { return Location{},ErrDriverNotFound }
    if err != nil { return Location{},err }
    // Late heartbeats may update location, but cannot revive expired presence.
    if _,err := tx.ExecContext(ctx, `UPDATE driver_profiles p SET is_online=FALSE
        WHERE p.user_id=$1 AND p.is_online AND NOT EXISTS (
            SELECT 1 FROM driver_locations l WHERE l.driver_user_id=p.user_id
            AND l.updated_at BETWEEN statement_timestamp()-($2 * INTERVAL '1 second') AND statement_timestamp())`,driverUserID,driver.MarketplaceLocationMaxAge.Seconds());err!=nil { return Location{},err }
    var location Location
	err = tx.QueryRowContext(ctx, `
		INSERT INTO driver_locations (driver_user_id, latitude, longitude, updated_at)
		SELECT dp.user_id, $2, $3, NOW()
		FROM driver_profiles dp
		WHERE dp.user_id = $1
		  AND dp.status IN ('approved', 'active')
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
	if err:=tx.Commit();err!=nil { return Location{},err }
	return location, nil
}
