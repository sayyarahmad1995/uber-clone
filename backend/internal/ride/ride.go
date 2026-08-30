package ride

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidLocation = errors.New("invalid ride location")

type Status string

const StatusRequested Status = "requested"

type Location struct {
	Latitude  float64
	Longitude float64
}

type Request struct {
	ID          uuid.UUID
	RiderUserID uuid.UUID
	Pickup      Location
	Destination Location
	Status      Status
	CreatedAt   time.Time
}

type Repository interface {
	Create(ctx context.Context, riderUserID uuid.UUID, pickup, destination Location) (Request, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Create(ctx context.Context, riderUserID uuid.UUID, pickup, destination Location) (Request, error) {
	if !validLocation(pickup) || !validLocation(destination) {
		return Request{}, ErrInvalidLocation
	}
	return s.repository.Create(ctx, riderUserID, pickup, destination)
}

func validLocation(location Location) bool {
	return location.Latitude >= -90 && location.Latitude <= 90 && location.Longitude >= -180 && location.Longitude <= 180
}
