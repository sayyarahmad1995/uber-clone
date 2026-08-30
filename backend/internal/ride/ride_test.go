package ride

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	created Request
	pickup Location
	destination Location
}

func (f *fakeRepository) Create(_ context.Context, riderUserID uuid.UUID, pickup, destination Location) (Request, error) {
	f.pickup = pickup
	f.destination = destination
	f.created = Request{ID: uuid.New(), RiderUserID: riderUserID, Pickup: pickup, Destination: destination, Status: StatusRequested}
	return f.created, nil
}

func TestCreateRideRequest(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	riderID := uuid.New()
	pickup := Location{Latitude: 24.8607, Longitude: 67.0011}
	destination := Location{Latitude: 24.9056, Longitude: 67.0822}

	request, err := service.Create(context.Background(), riderID, pickup, destination)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if request.RiderUserID != riderID {
		t.Fatalf("expected rider %s, got %s", riderID, request.RiderUserID)
	}
	if request.Status != StatusRequested {
		t.Fatalf("expected requested status, got %q", request.Status)
	}
	if repo.pickup != pickup || repo.destination != destination {
		t.Fatalf("locations were not passed to repository")
	}
}

func TestCreateRejectsInvalidLocations(t *testing.T) {
	service := NewService(&fakeRepository{})
	tests := []struct {
		name string
		pickup Location
		destination Location
	}{
		{name: "pickup latitude", pickup: Location{Latitude: 91, Longitude: 0}, destination: Location{}},
		{name: "pickup longitude", pickup: Location{Latitude: 0, Longitude: 181}, destination: Location{}},
		{name: "destination latitude", pickup: Location{}, destination: Location{Latitude: -91, Longitude: 0}},
		{name: "destination longitude", pickup: Location{}, destination: Location{Latitude: 0, Longitude: -181}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), uuid.New(), tt.pickup, tt.destination)
			if !errors.Is(err, ErrInvalidLocation) {
				t.Fatalf("expected ErrInvalidLocation, got %v", err)
			}
		})
	}
}
