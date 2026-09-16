package offer

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRideNotFound       = errors.New("ride request not found")
	ErrRideNotOpen        = errors.New("ride request is not open for offers")
	ErrDriverIneligible   = errors.New("driver is not eligible to offer")
	ErrAmountOutOfRange   = errors.New("offer amount is outside allowed range")
	ErrOfferNotFound      = errors.New("ride offer not found")
	ErrOfferNotActionable = errors.New("ride offer is not actionable")
	ErrOpportunityNotOpen = errors.New("driver opportunity is not open")
)

const (
	MinimumPercent int64 = 90
	MaximumPercent int64 = 130
	DiscoveryLimit       = 50
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRejected Status = "rejected"
	StatusExpired  Status = "expired"
	StatusClosed   Status = "closed"
)

type Market struct {
	RideRequestID       uuid.UUID
	ProposedAmountMinor int64
	Currency            string
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type DiscoveryItem struct {
	PickupDistanceMeters float64
	RideRequestID        uuid.UUID
	RiderUserID          uuid.UUID
	Pickup               Location
	Destination          Location
	ProposedFare         Market
	CreatedAt            time.Time
	RideExpiresAt        time.Time
	OpportunityExpiresAt time.Time
	OwnOffer             *Offer
}

type Offer struct {
	RideRequestID uuid.UUID
	DriverUserID  uuid.UUID
	AmountMinor   int64
	Currency      string
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     time.Time
	DecidedAt     *time.Time
}

// RiderOffer is a comparison view; these values are not offer lifecycle state.
type RiderOffer struct {
	Offer
	Driver               *DriverSummary
	Vehicle              *VehicleSummary
	Service              *ServiceSummary
	PickupDistanceMeters *float64
	MatchesProposedFare  bool
	Selectable           bool
}

type DriverSummary struct {
	DisplayName string
}

type VehicleSummary struct {
	Make      string
	Model     string
	ModelYear int
	Color     string
}

type ServiceSummary struct {
	Code        string
	DisplayName string
}

type Submission struct {
	Offer Offer
}

type Repository interface {
	Market(context.Context, uuid.UUID) (Market, error)
	Discover(context.Context, uuid.UUID, int) ([]DiscoveryItem, error)
	Upsert(context.Context, uuid.UUID, uuid.UUID, int64, int64, int64, string) (Offer, error)
	Skip(context.Context, uuid.UUID, uuid.UUID) error
	ListForRider(context.Context, uuid.UUID, uuid.UUID) ([]RiderOffer, error)
	Reject(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Offer, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Offer, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return Service{repository: repository}
}

func (s Service) Discover(ctx context.Context, driverUserID uuid.UUID) ([]DiscoveryItem, error) {
	return s.repository.Discover(ctx, driverUserID, DiscoveryLimit)
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

	result, err := s.repository.Upsert(ctx, rideRequestID, driverUserID, amountMinor, minimum, maximum, market.Currency)
	if err != nil {
		return Submission{}, err
	}
	return Submission{Offer: result}, nil
}

func (s Service) AcceptProposed(ctx context.Context, rideRequestID, driverUserID uuid.UUID) (Submission, error) {
	market, err := s.repository.Market(ctx, rideRequestID)
	if err != nil {
		return Submission{}, err
	}
	minimum, maximum := Bounds(market.ProposedAmountMinor)
	result, err := s.repository.Upsert(ctx, rideRequestID, driverUserID, market.ProposedAmountMinor, minimum, maximum, market.Currency)
	if err != nil {
		return Submission{}, err
	}
	return Submission{Offer: result}, nil
}

func (s Service) Skip(ctx context.Context, rideRequestID, driverUserID uuid.UUID) error {
	return s.repository.Skip(ctx, rideRequestID, driverUserID)
}

func (s Service) ListForRider(ctx context.Context, rideRequestID, riderUserID uuid.UUID) ([]RiderOffer, error) {
	return s.repository.ListForRider(ctx, rideRequestID, riderUserID)
}

func (s Service) Reject(ctx context.Context, rideRequestID, riderUserID, driverUserID uuid.UUID) (Offer, error) {
	return s.repository.Reject(ctx, rideRequestID, riderUserID, driverUserID)
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
