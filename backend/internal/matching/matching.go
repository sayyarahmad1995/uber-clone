package matching

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRideNotFound      = errors.New("ride request not found")
	ErrRideNotMatchable  = errors.New("ride request does not use automatic matching")
	ErrRideNotOpen       = errors.New("ride request is not open for matching")
	ErrNoEligibleDriver  = errors.New("no eligible driver available")
	ErrCandidateNotFound = errors.New("driver candidate not found")
	ErrCandidateResolved = errors.New("driver candidate already resolved")
)

const (
	DefaultDriverLocationFreshness           = 2 * time.Minute
	DefaultAutomaticCandidateResponseTimeout = 30 * time.Second
)

type CandidateStatus string

const (
	CandidateStatusPending  CandidateStatus = "pending"
	CandidateStatusAccepted CandidateStatus = "accepted"
	CandidateStatusRejected CandidateStatus = "rejected"
)

type Candidate struct {
	RideRequestID uuid.UUID
	DriverUserID  uuid.UUID
	Status        CandidateStatus
	CreatedAt     time.Time
	DecidedAt     *time.Time
}

type Result struct {
	Candidate Candidate
	Created   bool
}

type Repository interface {
	Match(context.Context, uuid.UUID, uuid.UUID) (Result, error)
	Reject(context.Context, uuid.UUID, uuid.UUID) (Candidate, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Match(ctx context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	return s.repository.Match(ctx, rideRequestID, riderUserID)
}

func (s Service) Reject(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Candidate, error) {
	return s.repository.Reject(ctx, rideRequestID, driverUserID)
}
