package ride

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
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
	var request Request
	var amount sql.NullInt64
	var currency sql.NullString
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ride_requests (id,rider_user_id,pickup_latitude,pickup_longitude,destination_latitude,destination_longitude,proposed_fare_minor,currency,status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id,rider_user_id,pickup_latitude,pickup_longitude,destination_latitude,destination_longitude,proposed_fare_minor,currency,status,created_at
	`, uuid.New(), riderUserID, input.Pickup.Latitude, input.Pickup.Longitude, input.Destination.Latitude, input.Destination.Longitude, proposedAmount, proposedCurrency, StatusRequested).Scan(
		&request.ID, &request.RiderUserID, &request.Pickup.Latitude, &request.Pickup.Longitude, &request.Destination.Latitude, &request.Destination.Longitude, &amount, &currency, &request.Status, &request.CreatedAt)
	if err != nil {
		return Request{}, err
	}
	if amount.Valid && currency.Valid {
		request.ProposedFare = &Money{AmountMinor: amount.Int64, Currency: currency.String}
	}
	return request, nil
}
