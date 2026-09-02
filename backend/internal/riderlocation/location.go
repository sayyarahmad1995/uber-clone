package riderlocation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrActiveTripNotFound = errors.New("active trip not found")
	ErrLocationNotFound   = errors.New("driver location not found")
)

type View struct {
	RideRequestID uuid.UUID
	Latitude      float64
	Longitude     float64
	UpdatedAt     time.Time
}

type Repository interface {
	GetForRide(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (View, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) GetForRide(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (View, error) {
	return s.repository.GetForRide(ctx, rideRequestID, riderUserID)
}
