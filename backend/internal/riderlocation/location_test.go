package riderlocation

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	rideRequestID uuid.UUID
	riderUserID   uuid.UUID
	view          View
	err           error
}

func (r *fakeRepository) GetForRide(_ context.Context, rideRequestID, riderUserID uuid.UUID) (View, error) {
	r.rideRequestID = rideRequestID
	r.riderUserID = riderUserID
	return r.view, r.err
}

func TestGetForRideUsesRequestedRideAndAuthenticatedRiderIdentity(t *testing.T) {
	rideRequestID := uuid.New()
	riderUserID := uuid.New()
	expected := View{RideRequestID: rideRequestID, Latitude: 24.8607, Longitude: 67.0011, UpdatedAt: time.Now()}
	repository := &fakeRepository{view: expected}
	service := NewService(repository)

	view, err := service.GetForRide(context.Background(), rideRequestID, riderUserID)
	if err != nil {
		t.Fatalf("GetForRide returned error: %v", err)
	}
	if repository.rideRequestID != rideRequestID {
		t.Fatalf("repository ride request = %s, want %s", repository.rideRequestID, rideRequestID)
	}
	if repository.riderUserID != riderUserID {
		t.Fatalf("repository rider identity = %s, want %s", repository.riderUserID, riderUserID)
	}
	if view != expected {
		t.Fatalf("view = %#v, want %#v", view, expected)
	}
}
