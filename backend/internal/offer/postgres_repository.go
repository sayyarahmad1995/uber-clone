package offer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return PostgresRepository{db: db}
}

func (r PostgresRepository) Market(ctx context.Context, rideRequestID uuid.UUID) (Market, error) {
	var market Market
	var status string
	var proposedAmount sql.NullInt64
	var currency sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT rr.id, rr.proposed_fare_minor, rr.currency, rr.status
		FROM ride_requests rr
		WHERE rr.id = $1
		  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
	`, rideRequestID).Scan(
		&market.RideRequestID,
		&proposedAmount,
		&currency,
		&status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Market{}, ErrRideNotFound
	}
	if err != nil {
		return Market{}, err
	}
	if status != "requested" || !proposedAmount.Valid || !currency.Valid {
		return Market{}, ErrRideNotOpen
	}
	market.ProposedAmountMinor = proposedAmount.Int64
	market.Currency = currency.String
	return market, nil
}

func (r PostgresRepository) Upsert(ctx context.Context, rideRequestID, driverUserID uuid.UUID, amountMinor, minimumMinor, maximumMinor int64, currency string) (Offer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, err
	}
	defer tx.Rollback()

	var actualCurrency sql.NullString
	var proposedAmount sql.NullInt64
	var riderUserID uuid.UUID
	var status string
	if err := tx.QueryRowContext(ctx, `
		SELECT currency, proposed_fare_minor, rider_user_id, status
		FROM ride_requests
		WHERE id = $1
		FOR UPDATE
	`, rideRequestID).Scan(&actualCurrency, &proposedAmount, &riderUserID, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrRideNotFound
		}
		return Offer{}, err
	}
	if status != "requested" || !proposedAmount.Valid || !actualCurrency.Valid || actualCurrency.String != currency {
		return Offer{}, ErrRideNotOpen
	}
	if hasTrip, err := rideHasTrip(ctx, tx, rideRequestID); err != nil {
		return Offer{}, err
	} else if hasTrip {
		return Offer{}, ErrRideNotOpen
	}
	if driverUserID == riderUserID {
		return Offer{}, ErrDriverIneligible
	}
	if amountMinor < minimumMinor || amountMinor > maximumMinor {
		return Offer{}, ErrAmountOutOfRange
	}

	var eligible bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM driver_profiles p
			JOIN driver_vehicles v ON v.driver_user_id = p.user_id
			JOIN user_capabilities c ON c.user_id = p.user_id AND c.capability = 'driver'
			WHERE p.user_id = $1
			  AND p.status = 'active'
			  AND p.is_online = TRUE
			  AND NOT EXISTS (
				SELECT 1 FROM trips t
				WHERE t.driver_user_id = p.user_id
				  AND t.status IN ('assigned', 'in_progress')
			  )
		)
	`, driverUserID).Scan(&eligible); err != nil {
		return Offer{}, err
	}
	if !eligible {
		return Offer{}, ErrDriverIneligible
	}

	var result Offer
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ride_offers (ride_request_id, driver_user_id, amount_minor, currency, status, decided_at)
		VALUES ($1, $2, $3, $4, 'pending', NULL)
		ON CONFLICT (ride_request_id, driver_user_id)
		DO UPDATE SET amount_minor = EXCLUDED.amount_minor,
		              currency = EXCLUDED.currency,
		              status = 'pending',
		              decided_at = NULL,
		              updated_at = NOW()
		RETURNING ride_request_id, driver_user_id, amount_minor, currency, status, created_at, updated_at, decided_at
	`, rideRequestID, driverUserID, amountMinor, currency).Scan(
		&result.RideRequestID,
		&result.DriverUserID,
		&result.AmountMinor,
		&result.Currency,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DecidedAt,
	); err != nil {
		return Offer{}, err
	}

	if err := tx.Commit(); err != nil {
		return Offer{}, err
	}
	return result, nil
}

func (r PostgresRepository) ListForRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) ([]Offer, error) {
	var ownedID uuid.UUID
	var status string
	var proposedAmount sql.NullInt64
	var currency sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT rr.id, rr.status, rr.proposed_fare_minor, rr.currency
		FROM ride_requests rr
		WHERE rr.id = $1 AND rr.rider_user_id = $2
		  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)
	`, rideRequestID, riderUserID).Scan(&ownedID, &status, &proposedAmount, &currency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRideNotFound
		}
		return nil, err
	}
	if status != "requested" || !proposedAmount.Valid || !currency.Valid {
		return nil, ErrRideNotOpen
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT ride_request_id, driver_user_id, amount_minor, currency, status, created_at, updated_at, decided_at
		FROM ride_offers
		WHERE ride_request_id = $1
		ORDER BY amount_minor ASC, created_at ASC, driver_user_id ASC
	`, rideRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offers := make([]Offer, 0)
	for rows.Next() {
		var item Offer
		if err := rows.Scan(
			&item.RideRequestID,
			&item.DriverUserID,
			&item.AmountMinor,
			&item.Currency,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DecidedAt,
		); err != nil {
			return nil, err
		}
		offers = append(offers, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return offers, nil
}

func (r PostgresRepository) Reject(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID) (Offer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, err
	}
	defer tx.Rollback()

	var rideStatus string
	var proposedAmount sql.NullInt64
	var currency sql.NullString
	if err := tx.QueryRowContext(ctx, `
		SELECT status, proposed_fare_minor, currency
		FROM ride_requests
		WHERE id = $1 AND rider_user_id = $2
		FOR UPDATE
	`, rideRequestID, riderUserID).Scan(&rideStatus, &proposedAmount, &currency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrRideNotFound
		}
		return Offer{}, err
	}
	if rideStatus != "requested" || !proposedAmount.Valid || !currency.Valid {
		return Offer{}, ErrRideNotOpen
	}
	if hasTrip, err := rideHasTrip(ctx, tx, rideRequestID); err != nil {
		return Offer{}, err
	} else if hasTrip {
		return Offer{}, ErrRideNotOpen
	}

	result, err := getOfferForUpdate(ctx, tx, rideRequestID, driverUserID)
	if err != nil {
		return Offer{}, err
	}
	switch result.Status {
	case StatusPending:
		if err := tx.QueryRowContext(ctx, `
			UPDATE ride_offers
			SET status = 'rejected', decided_at = NOW(), updated_at = NOW()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING ride_request_id, driver_user_id, amount_minor, currency, status, created_at, updated_at, decided_at
		`, rideRequestID, driverUserID).Scan(
			&result.RideRequestID,
			&result.DriverUserID,
			&result.AmountMinor,
			&result.Currency,
			&result.Status,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.DecidedAt,
		); err != nil {
			return Offer{}, err
		}
	case StatusRejected:
		// Idempotent Rider rejection.
	default:
		return Offer{}, ErrOfferNotActionable
	}

	if err := tx.Commit(); err != nil {
		return Offer{}, err
	}
	return result, nil
}

func (r PostgresRepository) Get(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Offer, error) {
	var result Offer
	if err := r.db.QueryRowContext(ctx, `
		SELECT ride_request_id, driver_user_id, amount_minor, currency, status, created_at, updated_at, decided_at
		FROM ride_offers
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideRequestID, driverUserID).Scan(
		&result.RideRequestID,
		&result.DriverUserID,
		&result.AmountMinor,
		&result.Currency,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DecidedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrOfferNotFound
		}
		return Offer{}, err
	}
	return result, nil
}

func getOfferForUpdate(ctx context.Context, tx *sql.Tx, rideRequestID, driverUserID uuid.UUID) (Offer, error) {
	var result Offer
	if err := tx.QueryRowContext(ctx, `
		SELECT ride_request_id, driver_user_id, amount_minor, currency, status, created_at, updated_at, decided_at
		FROM ride_offers
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(
		&result.RideRequestID,
		&result.DriverUserID,
		&result.AmountMinor,
		&result.Currency,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DecidedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrOfferNotFound
		}
		return Offer{}, err
	}
	return result, nil
}

func rideHasTrip(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM trips WHERE ride_request_id = $1)`, rideRequestID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
