package trip

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAssignmentNotFound = errors.New("accepted assignment not found")
	ErrAssignmentResolved = errors.New("ride request candidate already resolved")
	ErrTripNotFound       = errors.New("trip not found")
	ErrTripNotStarted     = errors.New("trip has not started")
	ErrTripCompleted      = errors.New("trip already completed")
)

type Status string

const (
	StatusAssigned   Status = "assigned"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

type Trip struct {
	RideRequestID uuid.UUID
	RiderUserID   uuid.UUID
	DriverUserID  uuid.UUID
	Status        Status
	AssignedAt    time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

type Acceptance struct {
	Trip               Trip
	CandidateCreatedAt time.Time
	CandidateDecidedAt *time.Time
}

type Repository interface {
	Accept(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Acceptance, error)
	Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error)
	Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Accept(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Acceptance, error) {
	return s.repository.Accept(ctx, rideRequestID, driverUserID)
}

func (s Service) Start(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	return s.repository.Start(ctx, rideRequestID, driverUserID)
}

func (s Service) Complete(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	return s.repository.Complete(ctx, rideRequestID, driverUserID)
}
