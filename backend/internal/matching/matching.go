package matching

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRideNotFound     = errors.New("ride request not found")
	ErrNoEligibleDriver = errors.New("no eligible driver available")
)

type Candidate struct {
	RideRequestID uuid.UUID
	DriverUserID  uuid.UUID
	CreatedAt     time.Time
}

type Result struct {
	Candidate Candidate
	Created   bool
}

type Repository interface {
	Match(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Match(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	return s.repository.Match(ctx, rideRequestID, riderUserID)
}
