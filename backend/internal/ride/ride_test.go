package ride

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	created Request
	input   CreateInput
}

func (f *fakeRepository) Create(_ context.Context, riderUserID uuid.UUID, input CreateInput) (Request, error) {
	f.input = input
	f.created = Request{
		ID:           uuid.New(),
		RiderUserID:  riderUserID,
		Pickup:       input.Pickup,
		Destination:  input.Destination,
		BookingMode:  BookingModeOffers,
		ProposedFare: input.ProposedFare,
		Status:       StatusRequested,
	}
	return f.created, nil
}

func TestCreateRideRequestRequiresProposedFare(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Create(context.Background(), uuid.New(), CreateInput{
		Pickup:      Location{},
		Destination: Location{},
	})
	if !errors.Is(err, ErrInvalidFare) {
		t.Fatalf("expected ErrInvalidFare, got %v", err)
	}
}

func TestCreateRideRequestNormalizesFareCurrency(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	fare := Money{AmountMinor: 70000, Currency: " pkr "}

	request, err := service.Create(context.Background(), uuid.New(), CreateInput{
		Pickup:       Location{},
		Destination:  Location{},
		ProposedFare: &fare,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if request.ProposedFare == nil || request.ProposedFare.AmountMinor != 70000 || request.ProposedFare.Currency != "PKR" {
		t.Fatalf("unexpected proposed fare: %#v", request.ProposedFare)
	}
	if repo.input.ProposedFare == nil || repo.input.ProposedFare.Currency != "PKR" {
		t.Fatalf("repository received unnormalized fare: %#v", repo.input.ProposedFare)
	}
}

func TestCreateRejectsInvalidLocations(t *testing.T) {
	fare := Money{AmountMinor: 100, Currency: "PKR"}
	_, err := NewService(&fakeRepository{}).Create(context.Background(), uuid.New(), CreateInput{
		Pickup:       Location{Latitude: 91},
		Destination:  Location{},
		ProposedFare: &fare,
	})
	if !errors.Is(err, ErrInvalidLocation) {
		t.Fatalf("expected ErrInvalidLocation, got %v", err)
	}
}

func TestCreateRequiresValidFare(t *testing.T) {
	service := NewService(&fakeRepository{})
	tests := []*Money{
		nil,
		{AmountMinor: 0, Currency: "PKR"},
		{AmountMinor: 100, Currency: "US"},
	}

	for _, fare := range tests {
		_, err := service.Create(context.Background(), uuid.New(), CreateInput{
			Pickup:       Location{},
			Destination:  Location{},
			ProposedFare: fare,
		})
		if !errors.Is(err, ErrInvalidFare) {
			t.Fatalf("expected ErrInvalidFare for %#v, got %v", fare, err)
		}
	}
}
