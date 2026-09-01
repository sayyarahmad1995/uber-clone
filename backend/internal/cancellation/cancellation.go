package cancellation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

var (
	ErrNotFound      = errors.New("ride cancellation target not found")
	ErrTripCompleted = errors.New("completed trip cannot be cancelled")
)

type Result struct {
	RideRequestID uuid.UUID
	Status        ride.Status
	CancelledBy   ride.CancellationActor
	CancelledAt   time.Time
	Trip          *trip.Trip
}

type Repository interface {
	CancelByRider(context.Context, uuid.UUID, uuid.UUID) (Result, error)
	CancelByDriver(context.Context, uuid.UUID, uuid.UUID) (Result, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) CancelByRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	return s.repository.CancelByRider(ctx, rideRequestID, riderUserID)
}

func (s Service) CancelByDriver(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Result, error) {
	return s.repository.CancelByDriver(ctx, rideRequestID, driverUserID)
}
