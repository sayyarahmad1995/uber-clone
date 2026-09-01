package drivertrip

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

var ErrNotFound = errors.New("active trip not found")

const historyLimit = 50

type View struct {
	RideRequestID uuid.UUID
	Pickup        ride.Location
	Destination   ride.Location
	Status        trip.Status
	AssignedAt    time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	CancelledAt   *time.Time
}

type Repository interface {
	GetCurrent(ctx context.Context, driverUserID uuid.UUID) (View, error)
	ListHistory(ctx context.Context, driverUserID uuid.UUID, limit int) ([]View, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) GetCurrent(ctx context.Context, driverUserID uuid.UUID) (View, error) {
	return s.repository.GetCurrent(ctx, driverUserID)
}

func (s Service) ListHistory(ctx context.Context, driverUserID uuid.UUID) ([]View, error) {
	return s.repository.ListHistory(ctx, driverUserID, historyLimit)
}
