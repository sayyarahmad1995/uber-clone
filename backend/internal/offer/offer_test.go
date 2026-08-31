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
	upsertedAmount  int64
	upsertedMinimum int64
	upsertedMaximum int64
}

func (f *fakeRepository) Market(context.Context, uuid.UUID) (Market, error) {
	return f.market, nil
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

func (f *fakeRepository) ListForRider(context.Context, uuid.UUID, uuid.UUID) ([]Offer, error) {
	return nil, nil
}

func (f *fakeRepository) Reject(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Offer, error) {
	return Offer{}, nil
}

func (f *fakeRepository) Get(context.Context, uuid.UUID, uuid.UUID) (Offer, error) {
	return Offer{}, nil
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
