package offer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
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
		  AND rr.expires_at > statement_timestamp()
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

	if err := marketplace.Expire(ctx, tx); err != nil {
		return Offer{}, err
	}
	policy, err := marketplace.LoadTimingPolicy(ctx, tx)
	if err != nil {
		return Offer{}, err
	}

	var actualCurrency sql.NullString
	var proposedAmount sql.NullInt64
	var riderUserID uuid.UUID
	var status string
	var unexpired bool
	var serviceCode string
	if err := tx.QueryRowContext(ctx, `
		SELECT currency, proposed_fare_minor, rider_user_id, status, service_code,
		       expires_at > statement_timestamp()
		FROM ride_requests
		WHERE id = $1
		FOR UPDATE
	`, rideRequestID).Scan(&actualCurrency, &proposedAmount, &riderUserID, &status, &serviceCode, &unexpired); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrRideNotFound
		}
		return Offer{}, err
	}
	if status != "requested" || !unexpired || !proposedAmount.Valid || !actualCurrency.Valid || actualCurrency.String != currency {
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

	var opportunityStatus string
	var opportunityOpen bool
	if err := tx.QueryRowContext(ctx, `
		SELECT status, visible_until > statement_timestamp()
		FROM driver_ride_request_opportunities
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(&opportunityStatus, &opportunityOpen); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrOpportunityNotOpen
		}
		return Offer{}, err
	}
	if opportunityStatus != "open" || !opportunityOpen {
		return Offer{}, ErrOpportunityNotOpen
	}

	eligible, err := driver.LockMarketplaceEligible(ctx, tx, driverUserID, serviceCode)
	if err != nil {
		return Offer{}, err
	}
	if !eligible {
		return Offer{}, ErrDriverIneligible
	}
	var operation []byte
	err = tx.QueryRowContext(ctx, `
		SELECT jsonb_build_object('vehicle_id', v.id, 'service_code', rr.service_code,
		  'service_name', sc.display_name, 'driver_name', p.display_name,
		  'make', v.make, 'model', v.model, 'model_year', v.model_year,
		  'color', v.color, 'license_plate', v.license_plate,
		  'fare', jsonb_build_object('amount_minor', $3::bigint, 'currency', $4::text))
		FROM driver_operating_selections s
		JOIN driver_profiles p ON p.user_id=s.driver_user_id
		JOIN driver_vehicles v ON v.id=s.vehicle_id AND v.driver_user_id=p.user_id
		JOIN ride_requests rr ON rr.id=$1
		JOIN driver_vehicle_service_enrollments e ON e.vehicle_id=v.id AND e.service_code=rr.service_code
		JOIN driver_service_catalog sc ON sc.code=e.service_code AND sc.is_active
		WHERE s.driver_user_id=$2
	`, rideRequestID, driverUserID, amountMinor, currency).Scan(&operation)
	if errors.Is(err, sql.ErrNoRows) {
		return Offer{}, ErrDriverIneligible
	}
	if err != nil {
		return Offer{}, err
	}

	var result Offer
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO ride_offers (
			ride_request_id, driver_user_id, amount_minor, currency, status,
			decided_at, operation_context, expires_at
		)
		VALUES (
			$1, $2, $3, $4, 'pending', NULL, $5::jsonb,
			statement_timestamp() + ($6 * INTERVAL '1 second')
		)
		ON CONFLICT (ride_request_id, driver_user_id) DO NOTHING
		RETURNING ride_request_id, driver_user_id, amount_minor, currency, status,
			created_at, updated_at, expires_at, decided_at
	`, rideRequestID, driverUserID, amountMinor, currency, string(operation), policy.OfferDecisionTTLSeconds).Scan(
		&result.RideRequestID,
		&result.DriverUserID,
		&result.AmountMinor,
		&result.Currency,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.ExpiresAt,
		&result.DecidedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrOpportunityNotOpen
		}
		return Offer{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE driver_ride_request_opportunities
		SET status = 'offered', responded_at = statement_timestamp()
		WHERE ride_request_id = $1 AND driver_user_id = $2 AND status = 'open'
	`, rideRequestID, driverUserID); err != nil {
		return Offer{}, err
	}

	if err := tx.Commit(); err != nil {
		return Offer{}, err
	}
	return result, nil
}

func (r PostgresRepository) Skip(ctx context.Context, rideRequestID, driverUserID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := marketplace.Expire(ctx, tx); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE driver_ride_request_opportunities opportunity
		SET status = 'driver_skipped', responded_at = statement_timestamp()
		WHERE opportunity.ride_request_id = $1
		  AND opportunity.driver_user_id = $2
		  AND opportunity.status = 'open'
		  AND opportunity.visible_until > statement_timestamp()
		  AND EXISTS (
		      SELECT 1 FROM ride_requests rr
		      WHERE rr.id = opportunity.ride_request_id
		        AND rr.status = 'requested'
		        AND rr.expires_at > statement_timestamp()
		  )
	`, rideRequestID, driverUserID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrOpportunityNotOpen
	}
	return tx.Commit()
}

func (r PostgresRepository) Reject(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID) (Offer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, err
	}
	defer tx.Rollback()
	if err := marketplace.Expire(ctx, tx); err != nil {
		return Offer{}, err
	}

	var rideStatus string
	var proposedAmount sql.NullInt64
	var currency sql.NullString
	var unexpired bool
	if err := tx.QueryRowContext(ctx, `
		SELECT status, proposed_fare_minor, currency, expires_at > statement_timestamp()
		FROM ride_requests
		WHERE id = $1 AND rider_user_id = $2
		FOR UPDATE
	`, rideRequestID, riderUserID).Scan(&rideStatus, &proposedAmount, &currency, &unexpired); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Offer{}, ErrRideNotFound
		}
		return Offer{}, err
	}
	if rideStatus != "requested" || !unexpired || !proposedAmount.Valid || !currency.Valid {
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
			SET status = 'rejected', decided_at = statement_timestamp(), updated_at = statement_timestamp()
			WHERE ride_request_id = $1 AND driver_user_id = $2
			RETURNING ride_request_id, driver_user_id, amount_minor, currency, status,
				created_at, updated_at, expires_at, decided_at
		`, rideRequestID, driverUserID).Scan(
			&result.RideRequestID,
			&result.DriverUserID,
			&result.AmountMinor,
			&result.Currency,
			&result.Status,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.ExpiresAt,
			&result.DecidedAt,
		); err != nil {
			return Offer{}, err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE driver_ride_request_opportunities
			SET status = 'rider_rejected'
			WHERE ride_request_id = $1 AND driver_user_id = $2 AND status = 'offered'
		`, rideRequestID, driverUserID); err != nil {
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
		SELECT ride_request_id, driver_user_id, amount_minor, currency, status,
		       created_at, updated_at, expires_at, decided_at
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
		&result.ExpiresAt,
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
		SELECT ride_request_id, driver_user_id, amount_minor, currency, status,
		       created_at, updated_at, expires_at, decided_at
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
		&result.ExpiresAt,
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
