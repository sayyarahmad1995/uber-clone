package ridestatus

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

var ErrNotFound = errors.New("ride request status not found")

type View struct {
	RideRequest ride.Request
	Trip        *trip.Trip
}

type Repository interface {
	GetOwned(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (View, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) GetOwned(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (View, error) {
	return s.repository.GetOwned(ctx, rideRequestID, riderUserID)
}
