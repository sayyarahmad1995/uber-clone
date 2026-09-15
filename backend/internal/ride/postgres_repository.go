package ride

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) Create(ctx context.Context, riderUserID uuid.UUID, input CreateInput) (Request, error) {
	var proposedAmount any
	var proposedCurrency any
	if input.ProposedFare != nil {
		proposedAmount = input.ProposedFare.AmountMinor
		proposedCurrency = input.ProposedFare.Currency
	}
	policy, err := marketplace.LoadTimingPolicy(ctx, r.db)
	if err != nil {
		return Request{}, err
	}
	var request Request
	var amount sql.NullInt64
	var currency sql.NullString
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO ride_requests (
			id,rider_user_id,pickup_latitude,pickup_longitude,destination_latitude,destination_longitude,
			proposed_fare_minor,currency,status,service_code,expires_at
		)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,statement_timestamp() + ($11 * INTERVAL '1 second')
		WHERE EXISTS (SELECT 1 FROM driver_service_catalog WHERE code=$10 AND is_active)
		RETURNING id,rider_user_id,pickup_latitude,pickup_longitude,destination_latitude,destination_longitude,
			proposed_fare_minor,currency,status,created_at,expires_at
	`, uuid.New(), riderUserID, input.Pickup.Latitude, input.Pickup.Longitude, input.Destination.Latitude, input.Destination.Longitude, proposedAmount, proposedCurrency, StatusRequested, input.ServiceCode, policy.RideRequestTTLSeconds).Scan(
		&request.ID,
		&request.RiderUserID,
		&request.Pickup.Latitude,
		&request.Pickup.Longitude,
		&request.Destination.Latitude,
		&request.Destination.Longitude,
		&amount,
		&currency,
		&request.Status,
		&request.CreatedAt,
		&request.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return Request{}, ErrInvalidService
	}
	if err != nil {
		return Request{}, err
	}
	if amount.Valid && currency.Valid {
		request.ProposedFare = &Money{AmountMinor: amount.Int64, Currency: currency.String}
	}
	request.ServiceCode = input.ServiceCode
	return request, nil
}
