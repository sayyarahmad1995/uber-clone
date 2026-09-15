package trip

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return PostgresRepository{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

const tripColumns = `
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

func scanOperationContext(raw []byte) (*marketplace.OperationContext, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var operationContext marketplace.OperationContext
	if err := json.Unmarshal(raw, &operationContext); err != nil {
		return nil, err
	}
	return &operationContext, nil
}

func scanTrip(row scanner) (Trip, error) {
	var result Trip
	var operationContextRaw []byte
	var settlementStatus string
	var settlementMethod sql.NullString
	if err := row.Scan(
		&result.RideRequestID,
		&result.RiderUserID,
		&result.DriverUserID,
		&result.Status,
		&result.AssignedAt,
		&result.StartedAt,
		&result.CompletedAt,
		&operationContextRaw,
		&result.CancelledAt,
		&settlementStatus,
		&settlementMethod,
		&result.Settlement.CashCollectedAt,
	); err != nil {
		return Trip{}, err
	}
	operationContext, err := scanOperationContext(operationContextRaw)
	if err != nil {
		return Trip{}, err
	}
	result.OperationContext = operationContext
	result.Settlement.Status = SettlementStatus(settlementStatus)
	if result.Settlement.Status == "" {
		result.Settlement.Status = SettlementUnsettled
	}
	if settlementMethod.Valid {
		method := settlementMethod.String
		result.Settlement.Method = &method
	}
	return result, nil
}

func (r PostgresRepository) Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()
	result, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}
	switch result.Status {
	case StatusAssigned:
		result, err = scanTrip(tx.QueryRowContext(ctx, `UPDATE trips SET status = 'in_progress', started_at = NOW() WHERE ride_request_id = $1 AND driver_user_id = $2 RETURNING `+tripColumns, rideRequestID, driverUserID))
		if err != nil {
			return Trip{}, err
		}
	case StatusInProgress:
	case StatusCompleted:
		return Trip{}, ErrTripCompleted
	case StatusCancelled:
		return Trip{}, ErrTripCancelled
	default:
		return Trip{}, errors.New("unknown trip status")
	}
	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return result, nil
}

func (r PostgresRepository) Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()
	result, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}
	switch result.Status {
	case StatusAssigned:
		return Trip{}, ErrTripNotStarted
	case StatusInProgress:
		result, err = scanTrip(tx.QueryRowContext(ctx, `UPDATE trips SET status = 'completed', completed_at = NOW() WHERE ride_request_id = $1 AND driver_user_id = $2 RETURNING `+tripColumns, rideRequestID, driverUserID))
		if err != nil {
			return Trip{}, err
		}
	case StatusCompleted:
	case StatusCancelled:
		return Trip{}, ErrTripCancelled
	default:
		return Trip{}, errors.New("unknown trip status")
	}
	if result.CompletedAt == nil {
		return Trip{}, errors.New("completed trip missing completed_at")
	}
	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return result, nil
}

func (r PostgresRepository) ConfirmCashCollected(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Trip{}, err
	}
	defer tx.Rollback()
	result, err := selectTrip(ctx, tx, rideRequestID, driverUserID, true)
	if err != nil {
		return Trip{}, err
	}
	switch result.Status {
	case StatusAssigned, StatusInProgress:
		return Trip{}, ErrTripNotCompleted
	case StatusCompleted:
		if result.Settlement.Status != SettlementCashCollected {
			result, err = scanTrip(tx.QueryRowContext(ctx, `UPDATE trips SET settlement_status = 'cash_collected', settlement_method = 'cash', cash_collected_at = NOW(), cash_collected_by = driver_user_id WHERE ride_request_id = $1 AND driver_user_id = $2 RETURNING `+tripColumns, rideRequestID, driverUserID))
			if err != nil {
				return Trip{}, err
			}
		}
	case StatusCancelled:
		return Trip{}, ErrTripCancelled
	default:
		return Trip{}, errors.New("unknown trip status")
	}
	if result.Settlement.Status != SettlementCashCollected || result.Settlement.Method == nil || result.Settlement.CashCollectedAt == nil {
		return Trip{}, errors.New("cash settlement missing persisted confirmation")
	}
	if err := tx.Commit(); err != nil {
		return Trip{}, err
	}
	return result, nil
}

func selectTrip(ctx context.Context, tx *sql.Tx, rideRequestID, driverUserID uuid.UUID, forUpdate bool) (Trip, error) {
	query := `SELECT ` + tripColumns + ` FROM trips WHERE ride_request_id = $1 AND driver_user_id = $2`
	if forUpdate {
		query += " FOR UPDATE"
	}
	result, err := scanTrip(tx.QueryRowContext(ctx, query, rideRequestID, driverUserID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Trip{}, ErrTripNotFound
		}
		return Trip{}, err
	}
	return result, nil
}
