package trip

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTripNotFound     = errors.New("trip not found")
	ErrTripNotStarted   = errors.New("trip has not started")
	ErrTripNotCompleted = errors.New("trip is not completed")
	ErrTripCompleted    = errors.New("trip already completed")
	ErrTripCancelled    = errors.New("trip already cancelled")
)

type Status string

const (
	StatusAssigned   Status = "assigned"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type SettlementStatus string

const (
	SettlementUnsettled     SettlementStatus = "unsettled"
	SettlementCashCollected SettlementStatus = "cash_collected"
)

type Settlement struct {
	Status          SettlementStatus
	Method          *string
	CashCollectedAt *time.Time
}

type OperationContext struct {
	DriverName   string    `json:"driver_name"`
	VehicleID    uuid.UUID `json:"vehicle_id"`
	Make         string    `json:"make"`
	Model        string    `json:"model"`
	ModelYear    int       `json:"model_year"`
	Color        string    `json:"color"`
	LicensePlate string    `json:"license_plate"`
	ServiceCode  string    `json:"service_code"`
	ServiceName  string    `json:"service_name"`
	Fare         Fare      `json:"fare"`
}

type Fare struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type Trip struct {
	OperationContext *OperationContext
	RideRequestID    uuid.UUID
	RiderUserID      uuid.UUID
	DriverUserID     uuid.UUID
	Status           Status
	AssignedAt       time.Time
	StartedAt        *time.Time
	CompletedAt      *time.Time
	CancelledAt      *time.Time
	Settlement       Settlement
}

type Repository interface {
	Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error)
	Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error)
	ConfirmCashCollected(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	return s.repository.Start(ctx, rideRequestID, driverUserID)
}

func (s Service) Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	return s.repository.Complete(ctx, rideRequestID, driverUserID)
}

func (s Service) ConfirmCashCollected(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	return s.repository.ConfirmCashCollected(ctx, rideRequestID, driverUserID)
}
