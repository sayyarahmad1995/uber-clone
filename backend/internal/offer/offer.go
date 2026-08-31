package offer

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

var (
	ErrRideNotFound       = errors.New("ride request not found")
	ErrRideNotOpen        = errors.New("ride request is not open for offers")
	ErrDriverIneligible   = errors.New("driver is not eligible to offer")
	ErrAmountOutOfRange   = errors.New("offer amount is outside allowed range")
	ErrOfferNotFound      = errors.New("ride offer not found")
	ErrOfferNotActionable = errors.New("ride offer is not actionable")
)

const (
	MinimumPercent int64 = 90
	MaximumPercent int64 = 130
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRejected Status = "rejected"
	StatusClosed   Status = "closed"
)

type Market struct {
	RideRequestID       uuid.UUID
	ProposedAmountMinor int64
	Currency            string
}

type Offer struct {
	RideRequestID uuid.UUID
	DriverUserID  uuid.UUID
	AmountMinor   int64
	Currency      string
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DecidedAt     *time.Time
}

type Submission struct {
	Offer Offer
	Trip  *trip.Trip
}

type Repository interface {
	Market(context.Context, uuid.UUID) (Market, error)
	Upsert(context.Context, uuid.UUID, uuid.UUID, int64, int64, int64, string) (Offer, error)
	ListForRider(context.Context, uuid.UUID, uuid.UUID) ([]Offer, error)
	Reject(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Offer, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Offer, error)
}

type Service struct {
	repository Repository
	trips      trip.Service
}

func NewService(repository Repository, trips trip.Service) Service {
	return Service{repository: repository, trips: trips}
}

func (s Service) Submit(ctx context.Context, rideRequestID, driverUserID uuid.UUID, amountMinor int64) (Submission, error) {
	market, err := s.repository.Market(ctx, rideRequestID)
	if err != nil {
		return Submission{}, err
	}
	minimum, maximum := Bounds(market.ProposedAmountMinor)
	if amountMinor < minimum || amountMinor > maximum {
		return Submission{}, ErrAmountOutOfRange
	}

	if amountMinor == market.ProposedAmountMinor {
		return s.AcceptProposed(ctx, rideRequestID, driverUserID)
	}

	result, err := s.repository.Upsert(ctx, rideRequestID, driverUserID, amountMinor, minimum, maximum, market.Currency)
	if err != nil {
		return Submission{}, err
	}
	return Submission{Offer: result}, nil
}

func (s Service) AcceptProposed(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Submission, error) {
	assignedTrip, err := s.trips.AcceptProposedFare(ctx, rideRequestID, driverUserID)
	if err != nil {
		return Submission{}, mapTripAssignmentError(err)
	}
	acceptedOffer, err := s.repository.Get(ctx, rideRequestID, driverUserID)
	if err != nil {
		return Submission{}, err
	}
	return Submission{Offer: acceptedOffer, Trip: &assignedTrip}, nil
}

func (s Service) ListForRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) ([]Offer, error) {
	return s.repository.ListForRider(ctx, rideRequestID, riderUserID)
}

func (s Service) Accept(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID) (trip.Trip, error) {
	result, err := s.trips.SelectOffer(ctx, rideRequestID, riderUserID, driverUserID)
	if err != nil {
		return trip.Trip{}, mapTripAssignmentError(err)
	}
	return result, nil
}

func (s Service) Reject(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID) (Offer, error) {
	return s.repository.Reject(ctx, rideRequestID, riderUserID, driverUserID)
}

func mapTripAssignmentError(err error) error {
	switch {
	case errors.Is(err, trip.ErrMarketplaceNotOpen):
		return ErrRideNotOpen
	case errors.Is(err, trip.ErrMarketplaceOfferGone):
		return ErrOfferNotActionable
	case errors.Is(err, trip.ErrDriverUnavailable):
		return ErrDriverIneligible
	default:
		return err
	}
}

func Bounds(proposedAmountMinor int64) (int64, int64) {
	return percentCeil(proposedAmountMinor, MinimumPercent), percentFloor(proposedAmountMinor, MaximumPercent)
}

func percentFloor(value, percent int64) int64 {
	return (value/100)*percent + ((value%100)*percent)/100
}

func percentCeil(value, percent int64) int64 {
	whole := (value / 100) * percent
	fraction := (value % 100) * percent
	return whole + (fraction+99)/100
}
