package offer

import (
	"context"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

func (r PostgresRepository) Discover(ctx context.Context, driverUserID uuid.UUID, limit int) ([]DiscoveryItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := marketplace.Expire(ctx, tx); err != nil {
		return nil, err
	}
	policy, err := marketplace.LoadTimingPolicy(ctx, tx)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		WITH eligible AS (
			SELECT rr.id
			FROM ride_requests rr
			JOIN driver_profiles p ON p.user_id = $1 AND p.status = 'active' AND p.is_online
			JOIN driver_operating_selections s ON s.driver_user_id = p.user_id
			JOIN driver_vehicles v ON v.id = s.vehicle_id AND v.driver_user_id = p.user_id
			JOIN driver_vehicle_service_enrollments e ON e.vehicle_id = v.id AND e.service_code = rr.service_code
			JOIN driver_service_catalog sc ON sc.code = e.service_code AND sc.is_active
			JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
			JOIN driver_locations l ON l.driver_user_id = p.user_id
			LEFT JOIN driver_ride_request_opportunities existing
			  ON existing.ride_request_id = rr.id
			 AND existing.driver_user_id = $1
			WHERE rr.status = 'requested'
			  AND rr.expires_at > statement_timestamp()
			  AND l.updated_at BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp()
			  AND NOT EXISTS (SELECT 1 FROM trips active WHERE active.driver_user_id = p.user_id AND active.status IN ('assigned', 'in_progress'))
			  AND rr.proposed_fare_minor IS NOT NULL
			  AND rr.currency IS NOT NULL
			  AND rr.rider_user_id <> $1
			  AND existing.ride_request_id IS NULL
			  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
			ORDER BY `+pickupDistanceSQL+` ASC, rr.created_at ASC, rr.id ASC
			LIMIT GREATEST(
				$2 - (
					SELECT COUNT(*)::int
					FROM driver_ride_request_opportunities active_opportunity
					WHERE active_opportunity.driver_user_id = $1
					  AND active_opportunity.status = 'open'
					  AND active_opportunity.visible_until > statement_timestamp()
				),
				0
			)
		)
		INSERT INTO driver_ride_request_opportunities (
			ride_request_id, driver_user_id, status, opened_at, visible_until
		)
		SELECT id, $1, 'open', statement_timestamp(),
			statement_timestamp() + ($4 * INTERVAL '1 second')
		FROM eligible
		ON CONFLICT (ride_request_id, driver_user_id) DO NOTHING
	`, driverUserID, limit, driver.MarketplaceLocationMaxAge.Seconds(), policy.DriverOpportunityTTLSeconds)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			rr.id,
			rr.rider_user_id,
			rr.pickup_latitude,
			rr.pickup_longitude,
			rr.destination_latitude,
			rr.destination_longitude,
			rr.proposed_fare_minor,
			rr.currency,
			rr.created_at,
			rr.expires_at,
			opportunity.visible_until,
`+pickupDistanceSQL+` AS pickup_distance_meters
		FROM driver_ride_request_opportunities opportunity
		JOIN ride_requests rr ON rr.id = opportunity.ride_request_id
		JOIN driver_profiles p ON p.user_id = $1 AND p.status = 'active' AND p.is_online
		JOIN driver_operating_selections s ON s.driver_user_id = p.user_id
		JOIN driver_vehicles v ON v.id = s.vehicle_id AND v.driver_user_id = p.user_id
		JOIN driver_vehicle_service_enrollments e ON e.vehicle_id = v.id AND e.service_code = rr.service_code
		JOIN driver_service_catalog sc ON sc.code = e.service_code AND sc.is_active
		JOIN driver_locations l ON l.driver_user_id = p.user_id
		WHERE opportunity.driver_user_id = $1
		  AND opportunity.status = 'open'
		  AND opportunity.visible_until > statement_timestamp()
		  AND rr.status = 'requested'
		  AND rr.expires_at > statement_timestamp()
		  AND l.updated_at BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp()
		  AND NOT EXISTS (SELECT 1 FROM trips active WHERE active.driver_user_id = p.user_id AND active.status IN ('assigned', 'in_progress'))
		  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
		ORDER BY pickup_distance_meters ASC, rr.created_at ASC, rr.id ASC
		LIMIT $2
	`, driverUserID, limit, driver.MarketplaceLocationMaxAge.Seconds())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DiscoveryItem, 0)
	for rows.Next() {
		var item DiscoveryItem
		if err := rows.Scan(
			&item.RideRequestID,
			&item.RiderUserID,
			&item.Pickup.Latitude,
			&item.Pickup.Longitude,
			&item.Destination.Latitude,
			&item.Destination.Longitude,
			&item.ProposedFare.ProposedAmountMinor,
			&item.ProposedFare.Currency,
			&item.CreatedAt,
			&item.RideExpiresAt,
			&item.OpportunityExpiresAt,
			&item.PickupDistanceMeters,
		); err != nil {
			return nil, err
		}
		item.ProposedFare.RideRequestID = item.RideRequestID
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return items, nil
}
