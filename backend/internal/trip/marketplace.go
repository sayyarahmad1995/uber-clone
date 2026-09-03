package trip

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
)

func (r PostgresRepository) SelectOffer(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()

	var actualRider uuid.UUID
	var status string
	var proposedAmount sql.NullInt64
	var currency sql.NullString
	if err := tx.QueryRowContext(ctx, `
		SELECT rider_user_id, status, proposed_fare_minor, currency
		FROM ride_requests
		WHERE id = $1
		FOR UPDATE
	`, rideRequestID).Scan(&actualRider, &status, &proposedAmount, &currency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Trip{}, ErrMarketplaceNotOpen
		}
		return Trip{}, err
	}
	if actualRider != riderUserID {
		return Trip{}, ErrMarketplaceOfferGone
	}
	if status != "requested" || !proposedAmount.Valid || !currency.Valid {
		return Trip{}, ErrMarketplaceNotOpen
	}
	if _, found, err := selectTripByRide(ctx, tx, rideRequestID); err != nil {
		return Trip{}, err
	} else if found {
		return Trip{}, ErrMarketplaceNotOpen
	}

	var offerStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT status
		FROM ride_offers
		WHERE ride_request_id = $1 AND driver_user_id = $2
		FOR UPDATE
	`, rideRequestID, driverUserID).Scan(&offerStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Trip{}, ErrMarketplaceOfferGone
		}
		return Trip{}, err
	}
	if offerStatus != "pending" {
		return Trip{}, ErrMarketplaceOfferGone
	}

	if err := lockEligibleMarketplaceDriver(ctx, tx, driverUserID); err != nil {
		return Trip{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE ride_offers
		SET status = 'accepted', decided_at = NOW(), updated_at = NOW()
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideRequestID, driverUserID); err != nil {
		return Trip{}, err
	}

	trip, err := insertMarketplaceTrip(ctx, tx, rideRequestID, riderUserID, driverUserID)
	if err != nil {
		return Trip{}, err
	}
	if err := closeCompetingOffers(ctx, tx, rideRequestID, driverUserID); err != nil {
		return Trip{}, err
	}
	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return trip, nil
}

func lockEligibleMarketplaceDriver(ctx context.Context, tx *sql.Tx, driverUserID uuid.UUID) error {
	eligible, err := driver.LockMarketplaceEligible(ctx, tx, driverUserID)
	if err != nil {
		return err
	}
	if !eligible {
		return ErrDriverUnavailable
	}
	return nil
}
func insertMarketplaceTrip(ctx context.Context, tx *sql.Tx, rideRequestID, riderUserID, driverUserID uuid.UUID) (Trip, error) {
	var trip Trip
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO trips (ride_request_id, rider_user_id, driver_user_id, assigned_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at
	`, rideRequestID, riderUserID, driverUserID).Scan(
		&trip.RideRequestID,
		&trip.RiderUserID,
		&trip.DriverUserID,
		&trip.Status,
		&trip.AssignedAt,
		&trip.StartedAt,
		&trip.CompletedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "trips_active_driver_idx" {
			return Trip{}, ErrDriverUnavailable
		}
		return Trip{}, err
	}
	return trip, nil
}

func closeCompetingOffers(ctx context.Context, tx *sql.Tx, rideRequestID, selectedDriverID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE ride_offers
		SET status = 'closed', decided_at = NOW(), updated_at = NOW()
		WHERE ride_request_id = $1
		  AND driver_user_id <> $2
		  AND status = 'pending'
	`, rideRequestID, selectedDriverID)
	return err
}

func selectTripByRide(ctx context.Context, tx *sql.Tx, rideRequestID uuid.UUID) (Trip, bool, error) {
	var trip Trip
	err := tx.QueryRowContext(ctx, `
		SELECT ride_request_id, rider_user_id, driver_user_id, status, assigned_at, started_at, completed_at
		FROM trips
		WHERE ride_request_id = $1
	`, rideRequestID).Scan(
		&trip.RideRequestID,
		&trip.RiderUserID,
		&trip.DriverUserID,
		&trip.Status,
		&trip.AssignedAt,
		&trip.StartedAt,
		&trip.CompletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Trip{}, false, nil
	}
	if err != nil {
		return Trip{}, false, err
	}
	return trip, true, nil
}
