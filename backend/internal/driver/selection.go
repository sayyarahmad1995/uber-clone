package driver

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var ErrSelectionInvalid = errors.New("select an owned vehicle with at least one approved, available service")
var ErrSelectionLocked = errors.New("go offline and finish the active trip before changing selection")
var ErrLocationRequired = errors.New("publish a fresh location before going online")
var ErrTripActive = errors.New("an active trip prevents going online")

type OperatingSelection struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	Valid     bool      `json:"valid"`
}

type OperatingState struct {
	Selection *OperatingSelection `json:"selection"`
	CanChange bool                `json:"can_change"`
}

func (s Service) OperatingState(ctx context.Context, id uuid.UUID) (OperatingState, error) {
	return s.repository.OperatingState(ctx, id)
}

func (s Service) SelectOperation(ctx context.Context, id, vehicle uuid.UUID) (OperatingState, error) {
	if vehicle == uuid.Nil {
		return OperatingState{}, ErrSelectionInvalid
	}
	return s.repository.SelectOperation(ctx, id, vehicle)
}

func (r PostgresRepository) OperatingState(ctx context.Context, id uuid.UUID) (OperatingState, error) {
	if err := r.expirePresence(ctx, id); err != nil {
		return OperatingState{}, err
	}
	var state OperatingState
	var vehicle sql.NullString
	var valid bool
	err := r.db.QueryRowContext(ctx, `
        SELECT NOT p.is_online AND NOT EXISTS (
            SELECT 1 FROM trips t WHERE t.driver_user_id=p.user_id AND t.status IN ('assigned','in_progress')),
            s.vehicle_id,
            EXISTS (SELECT 1 FROM driver_vehicles v
                JOIN driver_vehicle_service_enrollments e ON e.vehicle_id=v.id
                JOIN driver_service_catalog c ON c.code=e.service_code AND c.is_active
                WHERE v.id=s.vehicle_id AND v.driver_user_id=p.user_id)
        FROM driver_profiles p LEFT JOIN driver_operating_selections s ON s.driver_user_id=p.user_id
        WHERE p.user_id=$1`, id).Scan(&state.CanChange, &vehicle, &valid)
	if errors.Is(err, sql.ErrNoRows) {
		return state, ErrNotFound
	}
	if err != nil {
		return state, err
	}
	if vehicle.Valid {
		parsed, err := uuid.Parse(vehicle.String)
		if err != nil {
			return state, err
		}
		state.Selection = &OperatingSelection{VehicleID: parsed, Valid: valid}
	}
	return state, nil
}

func (r PostgresRepository) SelectOperation(ctx context.Context, id, vehicle uuid.UUID) (OperatingState, error) {
	if err := r.expirePresence(ctx, id); err != nil {
		return OperatingState{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return OperatingState{}, err
	}
	defer tx.Rollback()
	var online bool
	err = tx.QueryRowContext(ctx, `SELECT is_online FROM driver_profiles WHERE user_id=$1 AND status IN ('approved','active') FOR UPDATE`, id).Scan(&online)
	if errors.Is(err, sql.ErrNoRows) {
		return OperatingState{}, ErrNotFound
	}
	if err != nil {
		return OperatingState{}, err
	}
	var trip bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM trips WHERE driver_user_id=$1 AND status IN ('assigned','in_progress'))`, id).Scan(&trip); err != nil {
		return OperatingState{}, err
	}
	if online || trip {
		return OperatingState{}, ErrSelectionLocked
	}
	var selected uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT v.id FROM driver_vehicles v
        JOIN driver_vehicle_service_enrollments e ON e.vehicle_id=v.id
        JOIN driver_service_catalog c ON c.code=e.service_code
        JOIN user_capabilities u ON u.user_id=v.driver_user_id AND u.capability='driver'
        WHERE v.id=$2 AND v.driver_user_id=$1 AND c.is_active
        FOR SHARE OF v,e,c,u
        LIMIT 1`, id, vehicle).Scan(&selected)
	if errors.Is(err, sql.ErrNoRows) {
		return OperatingState{}, ErrSelectionInvalid
	}
	if err != nil {
		return OperatingState{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO driver_operating_selections (driver_user_id,vehicle_id)
        VALUES ($1,$2) ON CONFLICT (driver_user_id) DO UPDATE
        SET vehicle_id=EXCLUDED.vehicle_id,updated_at=NOW()`, id, vehicle); err != nil {
		return OperatingState{}, err
	}
	if err := tx.Commit(); err != nil {
		return OperatingState{}, err
	}
	return r.OperatingState(ctx, id)
}
