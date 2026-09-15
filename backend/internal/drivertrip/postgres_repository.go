package drivertrip

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func decodeOperationContext(raw []byte) (*marketplace.OperationContext, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var operationContext marketplace.OperationContext
	if err := json.Unmarshal(raw, &operationContext); err != nil {
		return nil, err
	}
	return &operationContext, nil
}

func (r PostgresRepository) GetCurrent(ctx context.Context, driverUserID uuid.UUID) (View, error) {
	var view View
	var operationContextRaw []byte
	var settlementStatus string
	var settlementMethod sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT
			t.ride_request_id,
			rr.pickup_latitude,
			rr.pickup_longitude,
			rr.destination_latitude,
			rr.destination_longitude,
			t.status,
			t.assigned_at,
			t.started_at,
			COALESCE(t.operation_context, 'null'::jsonb),
			t.settlement_status,
			t.settlement_method,
			t.cash_collected_at
		FROM trips t
		JOIN ride_requests rr ON rr.id = t.ride_request_id
		WHERE t.driver_user_id = $1
		  AND (
		    t.status IN ('assigned', 'in_progress')
		    OR (t.status = 'completed' AND t.settlement_status = 'unsettled')
		  )
		ORDER BY t.assigned_at DESC, t.ride_request_id DESC
		LIMIT 1
	`, driverUserID).Scan(
		&view.RideRequestID,
		&view.Pickup.Latitude,
		&view.Pickup.Longitude,
		&view.Destination.Latitude,
		&view.Destination.Longitude,
		&view.Status,
		&view.AssignedAt,
		&view.StartedAt,
		&operationContextRaw,
		&settlementStatus,
		&settlementMethod,
		&view.Settlement.CashCollectedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return View{}, ErrNotFound
	}
	if err != nil {
		return View{}, err
	}
	operationContext, err := decodeOperationContext(operationContextRaw)
	if err != nil {
		return View{}, err
	}
	view.OperationContext = operationContext
	applySettlement(&view, settlementStatus, settlementMethod)
	return view, nil
}

func (r PostgresRepository) ListHistory(ctx context.Context, driverUserID uuid.UUID, limit int) ([]View, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			t.ride_request_id,
			rr.pickup_latitude,
			rr.pickup_longitude,
			rr.destination_latitude,
			rr.destination_longitude,
			t.status,
			t.assigned_at,
			t.started_at,
			t.completed_at,
			t.cancelled_at,
			COALESCE(t.operation_context, 'null'::jsonb),
			t.settlement_status,
			t.settlement_method,
			t.cash_collected_at
		FROM trips t
		JOIN ride_requests rr ON rr.id = t.ride_request_id
		WHERE t.driver_user_id = $1
		  AND t.status IN ('completed', 'cancelled')
		ORDER BY t.assigned_at DESC, t.ride_request_id DESC
		LIMIT $2
	`, driverUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views := make([]View, 0)
	for rows.Next() {
		var view View
		var operationContextRaw []byte
		var settlementStatus string
		var settlementMethod sql.NullString
		if err := rows.Scan(
			&view.RideRequestID,
			&view.Pickup.Latitude,
			&view.Pickup.Longitude,
			&view.Destination.Latitude,
			&view.Destination.Longitude,
			&view.Status,
			&view.AssignedAt,
			&view.StartedAt,
			&view.CompletedAt,
			&view.CancelledAt,
			&operationContextRaw,
			&settlementStatus,
			&settlementMethod,
			&view.Settlement.CashCollectedAt,
		); err != nil {
			return nil, err
		}
		operationContext, err := decodeOperationContext(operationContextRaw)
		if err != nil {
			return nil, err
		}
		view.OperationContext = operationContext
		applySettlement(&view, settlementStatus, settlementMethod)
		views = append(views, view)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return views, nil
}

func applySettlement(view *View, status string, method sql.NullString) {
	view.Settlement.Status = trip.SettlementStatus(status)
	if view.Settlement.Status == "" {
		view.Settlement.Status = trip.SettlementUnsettled
	}
	if method.Valid {
		value := method.String
		view.Settlement.Method = &value
	}
}
