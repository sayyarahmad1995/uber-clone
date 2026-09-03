package ride

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidLocation = errors.New("invalid ride location")
	ErrInvalidFare     = errors.New("invalid proposed fare")
)

const MaxFareMinor int64 = 1_000_000_000_000

type Status string
type CancellationActor string

const (
	StatusRequested Status = "requested"
	StatusCancelled Status = "cancelled"

	CancellationActorRider  CancellationActor = "rider"
	CancellationActorDriver CancellationActor = "driver"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

type Money struct {
	AmountMinor int64
	Currency    string
}

type CreateInput struct {
	Pickup       Location
	Destination  Location
	ProposedFare *Money
}

type Request struct {
	ID           uuid.UUID
	RiderUserID  uuid.UUID
	Pickup       Location
	Destination  Location
	ProposedFare *Money
	Status       Status
	CreatedAt    time.Time
	CancelledAt  *time.Time
	CancelledBy  CancellationActor
}

type Repository interface {
	Create(ctx context.Context, riderUserID uuid.UUID, input CreateInput) (Request, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Create(ctx context.Context, riderUserID uuid.UUID, input CreateInput) (Request, error) {
	if !validLocation(input.Pickup) || !validLocation(input.Destination) {
		return Request{}, ErrInvalidLocation
	}
	if input.ProposedFare == nil || !validMoney(*input.ProposedFare) {
		return Request{}, ErrInvalidFare
	}
	input.ProposedFare.Currency = strings.ToUpper(strings.TrimSpace(input.ProposedFare.Currency))
	return s.repository.Create(ctx, riderUserID, input)
}

func validLocation(location Location) bool {
	return location.Latitude >= -90 && location.Latitude <= 90 && location.Longitude >= -180 && location.Longitude <= 180
}

func validMoney(money Money) bool {
	currency := strings.ToUpper(strings.TrimSpace(money.Currency))
	if money.AmountMinor <= 0 || money.AmountMinor > MaxFareMinor || len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
