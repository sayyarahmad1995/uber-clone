package driver

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// LockMarketplaceEligible serializes commercial responses/assignment with
// Driver availability, vehicle, capability and location changes. The caller
// owns the transaction and locks the Ride Request first.
func LockMarketplaceEligible(ctx context.Context, tx *sql.Tx, driverID uuid.UUID) (bool, error) {
	var locked uuid.UUID
	err := tx.QueryRowContext(ctx, `
		SELECT p.user_id
		FROM driver_profiles p
		JOIN driver_operating_selections s ON s.driver_user_id = p.user_id
        JOIN driver_vehicles v ON v.id = s.vehicle_id AND v.driver_user_id = p.user_id
        JOIN driver_vehicle_service_enrollments e ON e.vehicle_id = v.id AND e.service_code = s.service_code
        JOIN driver_service_catalog sc ON sc.code = e.service_code AND sc.is_active
		JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
		JOIN driver_locations l ON l.driver_user_id = p.user_id
		WHERE p.user_id = $1
		FOR UPDATE OF p FOR SHARE OF v, c, l, s, e, sc
	`, driverID).Scan(&locked)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Use a new statement after any lock wait, not the transaction start time.
	// The active-Trip unique index remains the final concurrent assignment guard.
	var eligible bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM driver_profiles p
			JOIN driver_operating_selections s ON s.driver_user_id = p.user_id
        JOIN driver_vehicles v ON v.id = s.vehicle_id AND v.driver_user_id = p.user_id
        JOIN driver_vehicle_service_enrollments e ON e.vehicle_id = v.id AND e.service_code = s.service_code
        JOIN driver_service_catalog sc ON sc.code = e.service_code AND sc.is_active
			JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
			JOIN driver_locations l ON l.driver_user_id = p.user_id
			WHERE p.user_id = $1 AND p.status = 'active' AND p.is_online
			  AND l.updated_at BETWEEN statement_timestamp() - ($2 * INTERVAL '1 second') AND statement_timestamp()
			  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.driver_user_id = p.user_id AND t.status IN ('assigned', 'in_progress'))
		)
	`, driverID, MarketplaceLocationMaxAge.Seconds()).Scan(&eligible)
	return eligible, err
}
