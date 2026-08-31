package offer

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

func (r PostgresRepository) Discover(ctx context.Context, driverUserID uuid.UUID, limit int) ([]DiscoveryItem, error) {
	var eligible bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM driver_profiles p
			JOIN driver_vehicles v ON v.driver_user_id = p.user_id
			JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
			WHERE p.user_id = $1
			  AND p.status = 'active'
			  AND p.is_online = TRUE
			  AND NOT EXISTS (
				SELECT 1
				FROM ride_driver_candidates active
				WHERE active.driver_user_id = p.user_id
				  AND active.status IN ('pending', 'accepted')
				  AND active.released_at IS NULL
			  )
			  AND NOT EXISTS (
				SELECT 1
				FROM trips t
				WHERE t.driver_user_id = p.user_id
				  AND t.status IN ('assigned', 'in_progress')
			  )
		)
	`, driverUserID).Scan(&eligible); err != nil {
		return nil, err
	}
	if !eligible {
		return []DiscoveryItem{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
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
			o.driver_user_id,
			o.amount_minor,
			o.currency,
			o.status,
			o.created_at,
			o.updated_at,
			o.decided_at
		FROM ride_requests rr
		LEFT JOIN ride_offers o
		  ON o.ride_request_id = rr.id
		 AND o.driver_user_id = $1
		WHERE rr.booking_mode = 'offers'
		  AND rr.status = 'requested'
		  AND rr.rider_user_id <> $1
		  AND NOT EXISTS (
			SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id
		  )
		ORDER BY rr.created_at DESC, rr.id DESC
		LIMIT $2
	`, driverUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DiscoveryItem, 0)
	for rows.Next() {
		var item DiscoveryItem
		var offerDriver uuid.NullUUID
		var offerAmount sql.NullInt64
		var offerCurrency, offerStatus sql.NullString
		var offerCreatedAt, offerUpdatedAt, offerDecidedAt sql.NullTime
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
			&offerDriver,
			&offerAmount,
			&offerCurrency,
			&offerStatus,
			&offerCreatedAt,
			&offerUpdatedAt,
			&offerDecidedAt,
		); err != nil {
			return nil, err
		}
		item.ProposedFare.RideRequestID = item.RideRequestID
		if offerDriver.Valid {
			item.OwnOffer = &Offer{
				RideRequestID: item.RideRequestID,
				DriverUserID:  offerDriver.UUID,
				AmountMinor:   offerAmount.Int64,
				Currency:      offerCurrency.String,
				Status:        Status(offerStatus.String),
				CreatedAt:     offerCreatedAt.Time,
				UpdatedAt:     offerUpdatedAt.Time,
			}
			if offerDecidedAt.Valid {
				item.OwnOffer.DecidedAt = &offerDecidedAt.Time
			}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
