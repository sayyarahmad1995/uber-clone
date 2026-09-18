package offer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
)

func (r PostgresRepository) ListForRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) ([]RiderOffer, error) {
	var status string
	var amount sql.NullInt64
	var currency sql.NullString
	var unexpired bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT rr.status, rr.proposed_fare_minor, rr.currency,
		       rr.expires_at > statement_timestamp()
		FROM ride_requests rr
		WHERE rr.id = $1 AND rr.rider_user_id = $2
		  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
	`, rideRequestID, riderUserID).Scan(&status, &amount, &currency, &unexpired); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRideNotFound
		}
		return nil, err
	}
	if status != "requested" || !unexpired || !amount.Valid || !currency.Valid {
		return nil, ErrRideNotOpen
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT o.ride_request_id, o.driver_user_id, o.amount_minor, o.currency,
		       o.status, o.created_at, o.updated_at, o.expires_at, o.decided_at,
		       o.operation_context->>'driver_name',
		       o.operation_context->>'make', o.operation_context->>'model', (o.operation_context->>'model_year')::integer, o.operation_context->>'color',
		       o.operation_context->>'service_code', o.operation_context->>'service_name',
		       CASE WHEN l.updated_at BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp()
		            THEN `+pickupDistanceSQL+` END AS pickup_distance_meters,
		       o.amount_minor = rr.proposed_fare_minor AND o.currency = rr.currency AS matches_proposed_fare,
		       COALESCE(o.status = 'pending'
		         AND o.expires_at > statement_timestamp()
		         AND rr.expires_at > statement_timestamp()
		         AND opportunity.status = 'offered'
		         AND p.status = 'active' AND p.is_online
		         AND v.driver_user_id IS NOT NULL AND c.user_id IS NOT NULL
		         AND sc.code = rr.service_code
		         AND o.operation_context->>'vehicle_id' = s.vehicle_id::text
		         AND o.operation_context->>'service_code' = rr.service_code
		         AND l.updated_at BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp()
		         AND o.driver_user_id <> rr.rider_user_id
		         AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.driver_user_id = o.driver_user_id AND t.status IN ('assigned', 'in_progress')), FALSE) AS selectable
		FROM ride_requests rr
		JOIN ride_offers o ON o.ride_request_id = rr.id
		JOIN driver_ride_request_opportunities opportunity
		  ON opportunity.ride_request_id = o.ride_request_id
		 AND opportunity.driver_user_id = o.driver_user_id
		LEFT JOIN driver_profiles p ON p.user_id = o.driver_user_id
		LEFT JOIN driver_operating_selections s ON s.driver_user_id = p.user_id
		LEFT JOIN driver_vehicles v ON v.id = s.vehicle_id AND v.driver_user_id = p.user_id
		LEFT JOIN driver_vehicle_service_enrollments e ON e.vehicle_id = v.id AND e.service_code = rr.service_code
		LEFT JOIN driver_service_catalog sc ON sc.code = e.service_code AND sc.is_active
		LEFT JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
		LEFT JOIN driver_locations l ON l.driver_user_id = p.user_id
		WHERE rr.id = $1 AND rr.rider_user_id = $2 AND rr.status = 'requested'
		  AND rr.expires_at > statement_timestamp()
		  AND o.status = 'pending'
		  AND o.expires_at > statement_timestamp()
		  AND opportunity.status = 'offered'
		  AND rr.proposed_fare_minor IS NOT NULL AND rr.currency IS NOT NULL
		  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
		ORDER BY selectable DESC, o.amount_minor ASC, pickup_distance_meters ASC NULLS LAST,
		         o.created_at ASC, o.driver_user_id ASC
	`, rideRequestID, riderUserID, driver.MarketplaceLocationMaxAge.Seconds())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]RiderOffer, 0)
	for rows.Next() {
		var item RiderOffer
		var displayName, make, model, color, serviceCode, serviceName sql.NullString
		var modelYear sql.NullInt64
		var distance sql.NullFloat64
		if err := rows.Scan(
			&item.RideRequestID,
			&item.DriverUserID,
			&item.AmountMinor,
			&item.Currency,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ExpiresAt,
			&item.DecidedAt,
			&displayName,
			&make,
			&model,
			&modelYear,
			&color,
			&serviceCode,
			&serviceName,
			&distance,
			&item.MatchesProposedFare,
			&item.Selectable,
		); err != nil {
			return nil, err
		}
		if displayName.Valid {
			item.Driver = &DriverSummary{DisplayName: displayName.String}
		}
		if make.Valid && model.Valid && color.Valid {
			item.Vehicle = &VehicleSummary{Make: make.String, Model: model.String, Color: color.String}
			if modelYear.Valid {
				item.Vehicle.ModelYear = int(modelYear.Int64)
			}
		}
		if serviceCode.Valid {
			displayName := serviceName.String
			if !serviceName.Valid || displayName == "" {
				displayName = serviceCode.String
			}
			item.Service = &ServiceSummary{Code: serviceCode.String, DisplayName: displayName}
		}
		if distance.Valid {
			item.PickupDistanceMeters = &distance.Float64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
