package offer

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type fakeRepository struct {
	market          Market
	discovery       []DiscoveryItem
	discoveryLimit  int
	upsertedAmount  int64
	upsertedMinimum int64
	upsertedMaximum int64
}

func (f *fakeRepository) Market(context.Context, uuid.UUID) (Market, error) {
	return f.market, nil
}

func (f *fakeRepository) Discover(_ context.Context, _ uuid.UUID, limit int) ([]DiscoveryItem, error) {
	f.discoveryLimit = limit
	return f.discovery, nil
}

func (f *fakeRepository) Upsert(_ context.Context, rideRequestID, driverUserID uuid.UUID, amountMinor, minimumMinor, maximumMinor int64, currency string) (Offer, error) {
	f.upsertedAmount, f.upsertedMinimum, f.upsertedMaximum = amountMinor, minimumMinor, maximumMinor
	return Offer{
		RideRequestID: rideRequestID,
		DriverUserID:  driverUserID,
		AmountMinor:   amountMinor,
		Currency:      currency,
		Status:        StatusPending,
	}, nil
}

func (f *fakeRepository) ListForRider(context.Context, uuid.UUID, uuid.UUID) ([]RiderOffer, error) {
	return nil, nil
}

func (f *fakeRepository) Reject(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Offer, error) {
	return Offer{}, nil
}

func (f *fakeRepository) Get(context.Context, uuid.UUID, uuid.UUID) (Offer, error) {
	return Offer{}, nil
}

func TestDiscoverUsesBoundedMarketplaceFeed(t *testing.T) {
	expected := []DiscoveryItem{{RideRequestID: uuid.New()}}
	repo := &fakeRepository{discovery: expected}
	service := NewService(repo, trip.Service{})

	actual, err := service.Discover(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if repo.discoveryLimit != DiscoveryLimit {
		t.Fatalf("expected discovery limit %d, got %d", DiscoveryLimit, repo.discoveryLimit)
	}
	if len(actual) != 1 || actual[0].RideRequestID != expected[0].RideRequestID {
		t.Fatalf("unexpected discovery result: %#v", actual)
	}
}

func TestBoundsUseNinetyToOneHundredThirtyPercent(t *testing.T) {
	minimum, maximum := Bounds(10000)
	if minimum != 9000 || maximum != 13000 {
		t.Fatalf("unexpected bounds: %d-%d", minimum, maximum)
	}
}

func TestSubmitAcceptsBoundaryCounteroffers(t *testing.T) {
	for _, amount := range []int64{9000, 13000} {
		repo := &fakeRepository{
			market: Market{
				RideRequestID:       uuid.New(),
				ProposedAmountMinor: 10000,
				Currency:            "PKR",
			},
		}
		service := NewService(repo, trip.Service{})
		if _, err := service.Submit(context.Background(), repo.market.RideRequestID, uuid.New(), amount); err != nil {
			t.Fatalf("Submit(%d) returned error: %v", amount, err)
		}
	}
}

func TestAcceptProposedCreatesPendingOfferWithoutTrip(t *testing.T) {
	repo := &fakeRepository{market: Market{RideRequestID: uuid.New(), ProposedAmountMinor: 10000, Currency: "PKR"}}
	service := NewService(repo, trip.Service{})

	result, err := service.AcceptProposed(context.Background(), repo.market.RideRequestID, uuid.New())
	if err != nil {
		t.Fatalf("AcceptProposed returned error: %v", err)
	}
	if repo.upsertedAmount != 10000 {
		t.Fatalf("expected proposed fare 10000, got %d", repo.upsertedAmount)
	}
	if result.Offer.Status != StatusPending {
		t.Fatalf("expected pending offer, got %s", result.Offer.Status)
	}
	if result.Trip != nil {
		t.Fatal("exact-fare Driver response must not assign a Trip")
	}
}

func TestSubmitProposedFareUsesSamePendingOfferPath(t *testing.T) {
	repo := &fakeRepository{market: Market{RideRequestID: uuid.New(), ProposedAmountMinor: 10000, Currency: "PKR"}}
	service := NewService(repo, trip.Service{})

	result, err := service.Submit(context.Background(), repo.market.RideRequestID, uuid.New(), 10000)
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if result.Offer.AmountMinor != 10000 || result.Offer.Status != StatusPending || result.Trip != nil {
		t.Fatalf("unexpected exact-fare submission: %#v", result)
	}
}

func TestSubmitRejectsOutsideRange(t *testing.T) {
	repo := &fakeRepository{
		market: Market{
			RideRequestID:       uuid.New(),
			ProposedAmountMinor: 10000,
			Currency:            "PKR",
		},
	}
	service := NewService(repo, trip.Service{})
	for _, amount := range []int64{8999, 13001} {
		_, err := service.Submit(context.Background(), repo.market.RideRequestID, uuid.New(), amount)
		if !errors.Is(err, ErrAmountOutOfRange) {
			t.Fatalf("expected ErrAmountOutOfRange for %d, got %v", amount, err)
		}
	}
}

func TestBoundsRoundMinimumUpAndMaximumDown(t *testing.T) {
	minimum, maximum := Bounds(101)
	if minimum != 91 || maximum != 131 {
		t.Fatalf("unexpected rounded bounds: %d-%d", minimum, maximum)
	}
}
