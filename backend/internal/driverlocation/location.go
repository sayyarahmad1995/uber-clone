package driverlocation

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidLocation = errors.New("invalid driver location")
	ErrDriverNotFound  = errors.New("driver profile not found")
)

type Input struct {
	Latitude  float64
	Longitude float64
}

type Location struct {
	DriverUserID uuid.UUID
	Latitude     float64
	Longitude    float64
	UpdatedAt    time.Time
}

type Repository interface {
	UpsertCurrent(ctx context.Context, driverUserID uuid.UUID, input Input) (Location, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) SetCurrent(ctx context.Context, driverUserID uuid.UUID, input Input) (Location, error) {
	if !validCoordinate(input.Latitude, -90, 90) || !validCoordinate(input.Longitude, -180, 180) {
		return Location{}, ErrInvalidLocation
	}
	return s.repository.UpsertCurrent(ctx, driverUserID, input)
}

func validCoordinate(value, min, max float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= min && value <= max
}
