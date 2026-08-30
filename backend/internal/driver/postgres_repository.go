package driver

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) UpsertProfile(ctx context.Context, userID uuid.UUID, vehicle VehicleInput) (Profile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Profile{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO driver_profiles (user_id, status, is_online, created_at, updated_at)
		VALUES ($1, $2, FALSE, $3, $3)
		ON CONFLICT (user_id) DO UPDATE SET status = EXCLUDED.status, updated_at = EXCLUDED.updated_at
	`, userID, StatusActive, now); err != nil {
		return Profile{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO driver_vehicles (id, driver_user_id, make, model, color, license_plate, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (driver_user_id) DO UPDATE SET
			make = EXCLUDED.make,
			model = EXCLUDED.model,
			color = EXCLUDED.color,
			license_plate = EXCLUDED.license_plate,
			updated_at = EXCLUDED.updated_at
	`, uuid.New(), userID, vehicle.Make, vehicle.Model, vehicle.Color, vehicle.LicensePlate, now); err != nil {
		return Profile{}, err
	}

	if err := tx.Commit(); err != nil {
		return Profile{}, err
	}
	return r.FindByUserID(ctx, userID)
}

func (r PostgresRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (Profile, error) {
	var p Profile
	err := r.db.QueryRowContext(ctx, `
		SELECT p.user_id, p.status, p.is_online, p.created_at, p.updated_at,
		       v.id, v.make, v.model, v.color, v.license_plate
		FROM driver_profiles p
		JOIN driver_vehicles v ON v.driver_user_id = p.user_id
		WHERE p.user_id = $1
	`, userID).Scan(
		&p.UserID, &p.Status, &p.IsOnline, &p.CreatedAt, &p.UpdatedAt,
		&p.Vehicle.ID, &p.Vehicle.Make, &p.Vehicle.Model, &p.Vehicle.Color, &p.Vehicle.LicensePlate,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (r PostgresRepository) SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE driver_profiles
		SET is_online = $2, updated_at = NOW()
		WHERE user_id = $1 AND status = $3
	`, userID, online, StatusActive)
	if err != nil {
		return Profile{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Profile{}, err
	}
	if rows == 0 {
		return Profile{}, ErrNotFound
	}
	return r.FindByUserID(ctx, userID)
}
