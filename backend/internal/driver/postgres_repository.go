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

func (r PostgresRepository) UpsertProfile(ctx context.Context, userID uuid.UUID, input OnboardingInput) (Profile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Profile{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO driver_profiles (user_id, display_name, status, is_online, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, $4, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, userID, input.DisplayName, StatusActive, now); err != nil {
		return Profile{}, err
	}

	vehicle := input.Vehicle
	var vehicleID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM driver_vehicles
		WHERE driver_user_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT 1
	`, userID).Scan(&vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		vehicleID = uuid.New()
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO driver_vehicles (id, driver_user_id, make, model, model_year, color, license_plate, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		`, vehicleID, userID, vehicle.Make, vehicle.Model, vehicle.ModelYear, vehicle.Color, vehicle.LicensePlate, now); err != nil {
			return Profile{}, err
		}
	} else if err != nil {
		return Profile{}, err
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE driver_vehicles
			SET make = $2,
			    model = $3,
			    model_year = $4,
			    color = $5,
			    license_plate = $6,
			    updated_at = $7
			WHERE id = $1
		`, vehicleID, vehicle.Make, vehicle.Model, vehicle.ModelYear, vehicle.Color, vehicle.LicensePlate, now); err != nil {
			return Profile{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Profile{}, err
	}
	return r.FindByUserID(ctx, userID)
}

func (r PostgresRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (Profile, error) {
 if err := r.expirePresence(ctx,userID); err != nil { return Profile{}, err }
	var p Profile
	err := r.db.QueryRowContext(ctx, `
		SELECT p.user_id, COALESCE(p.display_name, ''), p.status, p.is_online, p.created_at, p.updated_at,
		       v.id, v.make, v.model, COALESCE(v.model_year, 0), v.color, v.license_plate
		FROM driver_profiles p
		JOIN driver_vehicles v ON v.driver_user_id = p.user_id
		WHERE p.user_id = $1
		ORDER BY v.created_at ASC, v.id ASC
		LIMIT 1
	`, userID).Scan(
		&p.UserID, &p.DisplayName, &p.Status, &p.IsOnline, &p.CreatedAt, &p.UpdatedAt,
		&p.Vehicle.ID, &p.Vehicle.Make, &p.Vehicle.Model, &p.Vehicle.ModelYear, &p.Vehicle.Color, &p.Vehicle.LicensePlate,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, err
	}
	return p, nil
}

func (r PostgresRepository) ListVehicles(ctx context.Context, userID uuid.UUID) ([]Vehicle, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT v.id, v.make, v.model, COALESCE(v.model_year, 0), v.color, v.license_plate,
		       e.service_code, c.display_name, e.approved_at, c.is_active
		FROM driver_vehicles v
		LEFT JOIN driver_vehicle_service_enrollments e ON e.vehicle_id = v.id
		LEFT JOIN driver_service_catalog c ON c.code = e.service_code
		WHERE v.driver_user_id = $1
		ORDER BY v.created_at ASC, v.id ASC, c.sort_order ASC, e.service_code ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vehicles := []Vehicle{}
	indices := map[uuid.UUID]int{}
	for rows.Next() {
		var vehicle Vehicle
		var code, name sql.NullString
		var approved sql.NullTime
		var active sql.NullBool
		if err := rows.Scan(
			&vehicle.ID,
			&vehicle.Make,
			&vehicle.Model,
			&vehicle.ModelYear,
			&vehicle.Color,
			&vehicle.LicensePlate,
			&code, &name, &approved, &active,
		); err != nil {
			return nil, err
		}
		index, exists := indices[vehicle.ID]
		if !exists {
			index = len(vehicles)
			indices[vehicle.ID] = index
			vehicle.Enrollments = []ServiceEnrollment{}
			vehicles = append(vehicles, vehicle)
		}
		if code.Valid {
			vehicles[index].Enrollments = append(vehicles[index].Enrollments, ServiceEnrollment{
				ServiceCode: code.String, DisplayName: name.String,
				ApprovedAt: approved.Time, ServiceActive: active.Bool,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (r PostgresRepository) SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil { return Profile{}, err }
    defer tx.Rollback()
    var id uuid.UUID
    err = tx.QueryRowContext(ctx, `SELECT user_id FROM driver_profiles WHERE user_id=$1 FOR UPDATE`, userID).Scan(&id)
    if errors.Is(err, sql.ErrNoRows) { return Profile{}, ErrNotFound }
    if err != nil { return Profile{}, err }
    if online {
        err = tx.QueryRowContext(ctx, `SELECT v.id FROM driver_profiles p
            JOIN driver_operating_selections s ON s.driver_user_id=p.user_id
            JOIN driver_vehicles v ON v.id=s.vehicle_id AND v.driver_user_id=p.user_id
            JOIN driver_vehicle_service_enrollments e ON e.vehicle_id=v.id AND e.service_code=s.service_code
            JOIN driver_service_catalog c ON c.code=e.service_code AND c.is_active
            JOIN user_capabilities u ON u.user_id=p.user_id AND u.capability='driver'
            WHERE p.user_id=$1 AND p.status IN ('approved','active')
            FOR SHARE OF s,v,e,c,u`, userID).Scan(&id)
        if errors.Is(err, sql.ErrNoRows) { return Profile{}, ErrSelectionInvalid }
        if err != nil { return Profile{}, err }
        var updated time.Time
        err = tx.QueryRowContext(ctx, `SELECT updated_at FROM driver_locations WHERE driver_user_id=$1 FOR SHARE`, userID).Scan(&updated)
        if errors.Is(err, sql.ErrNoRows) { return Profile{}, ErrLocationRequired }
        if err != nil { return Profile{}, err }
        var fresh, trip bool
        err = tx.QueryRowContext(ctx, `SELECT $2::timestamptz BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp(),
            EXISTS(SELECT 1 FROM trips WHERE driver_user_id=$1 AND status IN ('assigned','in_progress'))`, userID, updated, MarketplaceLocationMaxAge.Seconds()).Scan(&fresh,&trip)
        if err != nil { return Profile{}, err }
        if !fresh { return Profile{}, ErrLocationRequired }
        if trip { return Profile{}, ErrTripActive }
    }
    if _, err := tx.ExecContext(ctx, `UPDATE driver_profiles SET is_online=$2,
        status=CASE WHEN $2 THEN 'active' ELSE status END,updated_at=NOW() WHERE user_id=$1`,userID,online); err != nil { return Profile{}, err }
    if err := tx.Commit(); err != nil { return Profile{}, err }
    return r.FindByUserID(ctx,userID)
}
