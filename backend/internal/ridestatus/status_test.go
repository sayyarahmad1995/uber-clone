package ridestatus

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	rideRequestID uuid.UUID
	riderUserID   uuid.UUID
	view          View
	err           error
}

func (r *fakeRepository) GetOwned(_ context.Context, rideRequestID, riderUserID uuid.UUID) (View, error) {
	r.rideRequestID = rideRequestID
	r.riderUserID = riderUserID
	return r.view, r.err
}

func TestGetOwnedScopesQueryToRideAndRider(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	rideRequestID := uuid.New()
	riderUserID := uuid.New()

	if _, err := service.GetOwned(context.Background(), rideRequestID, riderUserID); err != nil {
		t.Fatalf("GetOwned returned error: %v", err)
	}
	if repository.rideRequestID != rideRequestID {
		t.Fatalf("ride request ID = %s, want %s", repository.rideRequestID, rideRequestID)
	}
	if repository.riderUserID != riderUserID {
		t.Fatalf("rider user ID = %s, want %s", repository.riderUserID, riderUserID)
	}
}
