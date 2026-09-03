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
	if err := r.db.QueryRowContext(ctx, `
		SELECT rr.status, rr.proposed_fare_minor, rr.currency
		FROM ride_requests rr
		WHERE rr.id = $1 AND rr.rider_user_id = $2
		  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
	`, rideRequestID, riderUserID).Scan(&status, &amount, &currency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRideNotFound
		}
		return nil, err
	}
	if status != "requested" || !amount.Valid || !currency.Valid {
		return nil, ErrRideNotOpen
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT o.ride_request_id, o.driver_user_id, o.amount_minor, o.currency,
		       o.status, o.created_at, o.updated_at, o.decided_at,
		       p.display_name,
		       v.make, v.model, v.model_year, v.color,
		       CASE WHEN l.updated_at BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp()
		            THEN `+pickupDistanceSQL+` END AS pickup_distance_meters,
		       o.amount_minor = rr.proposed_fare_minor AND o.currency = rr.currency AS matches_proposed_fare,
		       COALESCE(o.status = 'pending' AND p.status = 'active' AND p.is_online
		         AND v.driver_user_id IS NOT NULL AND c.user_id IS NOT NULL
		         AND l.updated_at BETWEEN statement_timestamp() - ($3 * INTERVAL '1 second') AND statement_timestamp()
		         AND o.driver_user_id <> rr.rider_user_id
		         AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.driver_user_id = o.driver_user_id AND t.status IN ('assigned', 'in_progress')), FALSE) AS selectable
		FROM ride_requests rr
		JOIN ride_offers o ON o.ride_request_id = rr.id
		LEFT JOIN driver_profiles p ON p.user_id = o.driver_user_id
		LEFT JOIN driver_vehicles v ON v.driver_user_id = p.user_id
		LEFT JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
		LEFT JOIN driver_locations l ON l.driver_user_id = p.user_id
		WHERE rr.id = $1 AND rr.rider_user_id = $2 AND rr.status = 'requested'
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
		var displayName, make, model, color sql.NullString
		var modelYear sql.NullInt64
		var distance sql.NullFloat64
		if err := rows.Scan(&item.RideRequestID, &item.DriverUserID, &item.AmountMinor, &item.Currency,
			&item.Status, &item.CreatedAt, &item.UpdatedAt, &item.DecidedAt,
			&displayName, &make, &model, &modelYear, &color, &distance, &item.MatchesProposedFare, &item.Selectable); err != nil {
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
		if distance.Valid {
			item.PickupDistanceMeters = &distance.Float64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
