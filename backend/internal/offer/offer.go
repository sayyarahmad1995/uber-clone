package offer

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRideNotFound      = errors.New("ride request not found")
	ErrRideNotOpen       = errors.New("ride request is not open for offers")
	ErrDriverIneligible  = errors.New("driver is not eligible to offer")
	ErrAmountOutOfRange  = errors.New("offer amount is outside allowed range")
)

const (
	MinimumPercent int64 = 90
	MaximumPercent int64 = 130
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository interface {
	Market(context.Context, uuid.UUID) (Market, error)
	Upsert(context.Context, uuid.UUID, uuid.UUID, int64, int64, int64, string) (Offer, error)
	ListForRider(context.Context, uuid.UUID, uuid.UUID) ([]Offer, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Submit(ctx context.Context, rideRequestID, driverUserID uuid.UUID, amountMinor int64) (Offer, error) {
	market, err := s.repository.Market(ctx, rideRequestID)
	if err != nil {
		return Offer{}, err
	}
	minimum, maximum := Bounds(market.ProposedAmountMinor)
	if amountMinor < minimum || amountMinor > maximum {
		return Offer{}, ErrAmountOutOfRange
	}
	return s.repository.Upsert(ctx, rideRequestID, driverUserID, amountMinor, minimum, maximum, market.Currency)
}

func (s Service) ListForRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) ([]Offer, error) {
	return s.repository.ListForRider(ctx, rideRequestID, riderUserID)
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
