package marketplace

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func (r PostgresAssignmentRepository) SelectOffer(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID, expectedVersion time.Time) (trip.Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return trip.Trip{}, err
	}
	defer tx.Rollback()
	if err := Expire(ctx, tx); err != nil {
		return trip.Trip{}, err
	}

	var actualRider uuid.UUID
	var status string
	var proposedAmount sql.NullInt64
	var currency sql.NullString
	var rideUnexpired bool
	var serviceCode string
	if err := tx.QueryRowContext(ctx, `
		SELECT rider_user_id, status, proposed_fare_minor, currency, service_code,
		       expires_at > statement_timestamp()
		FROM ride_requests
		WHERE id = $1
		FOR UPDATE
	`, rideRequestID).Scan(&actualRider, &status, &proposedAmount, &currency, &serviceCode, &rideUnexpired); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return trip.Trip{}, ErrNotOpen
		}
		return trip.Trip{}, err
	}
	if actualRider != riderUserID {
		return trip.Trip{}, ErrOfferNotActionable
	}
	if status != "requested" || !rideUnexpired || !proposedAmount.Valid || !currency.Valid {
		return trip.Trip{}, ErrNotOpen
	}
	exists, err := tripExistsForRide(ctx, tx, rideRequestID)
	if err != nil {
		return trip.Trip{}, err
	}
	if exists {
		return trip.Trip{}, ErrNotOpen
	}

	var offerStatus string
	var offerVersion time.Time
	var offerUnexpired bool
	if err := tx.QueryRowContext(ctx, `
		SELECT status, updated_at, expires_at > statement_timestamp()
		FROM ride_offers
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(&offerStatus, &offerVersion, &offerUnexpired); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return trip.Trip{}, ErrOfferNotActionable
		}
		return trip.Trip{}, err
	}
	if offerStatus != "pending" || !offerUnexpired || !offerVersion.Equal(expectedVersion) {
		return trip.Trip{}, ErrOfferNotActionable
	}

	var opportunityStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT status
		FROM driver_ride_request_opportunities
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(&opportunityStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return trip.Trip{}, ErrOfferNotActionable
		}
		return trip.Trip{}, err
	}
	if opportunityStatus != "offered" {
		return trip.Trip{}, ErrOfferNotActionable
	}

	if err := lockEligibleMarketplaceDriver(ctx, tx, driverUserID, serviceCode); err != nil {
		return trip.Trip{}, err
	}
	var operationMatches bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM ride_offers o
		JOIN driver_operating_selections s ON s.driver_user_id=o.driver_user_id
		JOIN ride_requests rr ON rr.id=o.ride_request_id
		JOIN driver_vehicle_service_enrollments e ON e.vehicle_id=s.vehicle_id AND e.service_code=rr.service_code
		JOIN driver_service_catalog sc ON sc.code=e.service_code AND sc.is_active
		WHERE o.ride_request_id=$1 AND o.driver_user_id=$2
		  AND o.operation_context->>'vehicle_id'=s.vehicle_id::text
		  AND o.operation_context->>'service_code'=rr.service_code
	)`, rideRequestID, driverUserID).Scan(&operationMatches); err != nil {
		return trip.Trip{}, err
	}
	if !operationMatches {
		return trip.Trip{}, ErrDriverUnavailable
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ride_offers
		SET status = 'accepted', decided_at = statement_timestamp(), updated_at = statement_timestamp()
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideRequestID, driverUserID); err != nil {
		return trip.Trip{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE driver_ride_request_opportunities
		SET status = 'accepted'
		WHERE ride_request_id = $1 AND driver_user_id = $2 AND status = 'offered'
	`, rideRequestID, driverUserID); err != nil {
		return trip.Trip{}, err
	}

	assignedTrip, err := insertMarketplaceTrip(ctx, tx, rideRequestID, riderUserID, driverUserID)
	if err != nil {
		return trip.Trip{}, err
	}
	if err := markRideAccepted(ctx, tx, rideRequestID); err != nil {
		return trip.Trip{}, err
	}
	if err := closeCompetingOffers(ctx, tx, rideRequestID, driverUserID); err != nil {
		return trip.Trip{}, err
	}
	if err := closeCompetingOpportunities(ctx, tx, rideRequestID, driverUserID); err != nil {
		return trip.Trip{}, err
	}
	if err := tx.Commit(); err != nil {
		return trip.Trip{}, err
	}
	return assignedTrip, nil
}

func lockEligibleMarketplaceDriver(ctx context.Context, tx *sql.Tx, driverUserID uuid.UUID, serviceCode string) error {
	eligible, err := driver.LockMarketplaceEligible(ctx, tx, driverUserID, serviceCode)
	if err != nil {
		return err
	}
	if !eligible {
		return ErrDriverUnavailable
	}
	return nil
}

const assignedTripColumns = `
	ride_request_id,
	rider_user_id,
	driver_user_id,
	status,
	assigned_at,
	started_at,
	completed_at,
	COALESCE(operation_context, 'null'::jsonb),
	cancelled_at,
	settlement_status,
	settlement_method,
	cash_collected_at`

func scanAssignedTrip(row interface{ Scan(dest ...any) error }) (trip.Trip, error) {
	var result trip.Trip
	var settlementStatus string
	var settlementMethod sql.NullString
	var operationContext []byte
	if err := row.Scan(
		&result.RideRequestID,
		&result.RiderUserID,
		&result.DriverUserID,
		&result.Status,
		&result.AssignedAt,
		&result.StartedAt,
		&result.CompletedAt,
		&operationContext,
		&result.CancelledAt,
		&settlementStatus,
		&settlementMethod,
		&result.Settlement.CashCollectedAt,
	); err != nil {
		return trip.Trip{}, err
	}
	result.Settlement.Status = trip.SettlementStatus(settlementStatus)
	if string(operationContext) != "" && string(operationContext) != "null" {
		result.OperationContext = new(trip.OperationContext)
		if err := json.Unmarshal(operationContext, result.OperationContext); err != nil {
			return trip.Trip{}, err
		}
	}
	if result.Settlement.Status == "" {
		result.Settlement.Status = trip.SettlementUnsettled
	}
	if settlementMethod.Valid {
		method := settlementMethod.String
		result.Settlement.Method = &method
	}
	return result, nil
}

func insertMarketplaceTrip(ctx context.Context, tx *sql.Tx, rideRequestID, riderUserID, driverUserID uuid.UUID) (trip.Trip, error) {
	result, err := scanAssignedTrip(tx.QueryRowContext(ctx, `
		INSERT INTO trips (ride_request_id, rider_user_id, driver_user_id, assigned_at, operation_context)
		SELECT $1, $2, $3, statement_timestamp(), operation_context FROM ride_offers
		WHERE ride_request_id=$1 AND driver_user_id=$3
		RETURNING `+assignedTripColumns,
		rideRequestID,
		riderUserID,
		driverUserID,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "trips_active_driver_idx" {
			return trip.Trip{}, ErrDriverUnavailable
		}
		return trip.Trip{}, err
	}
	return result, nil
}

func markRideAccepted(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE ride_requests
		SET status = 'accepted'
		WHERE id = $1 AND status = 'requested' AND expires_at > statement_timestamp()
	`, rideRequestID)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated != 1 {
		return ErrNotOpen
	}
	return nil
}

func closeCompetingOffers(ctx context.Context, tx *sql.Tx, rideRequestID, selectedDriverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE ride_offers
		SET status = 'closed', decided_at = statement_timestamp(), updated_at = statement_timestamp()
		WHERE ride_request_id = $1
		  AND driver_user_id <> $2
		  AND status = 'pending'
	`, rideRequestID, selectedDriverID)
	return err
}

func closeCompetingOpportunities(ctx context.Context, tx *sql.Tx, rideRequestID, selectedDriverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE driver_ride_request_opportunities
		SET status = 'closed', responded_at = COALESCE(responded_at, statement_timestamp())
		WHERE ride_request_id = $1
		  AND driver_user_id <> $2
		  AND status IN ('open', 'offered')
	`, rideRequestID, selectedDriverID)
	return err
}

func tripExistsForRide(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM trips
			WHERE ride_request_id = $1
		)
	`, rideRequestID).Scan(&exists)
	return exists, err
}
