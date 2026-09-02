package riderlocation

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	riderUserID uuid.UUID
	view        View
	err         error
}

func (r *fakeRepository) GetForActiveTrip(_ context.Context, riderUserID uuid.UUID) (View, error) {
	r.riderUserID = riderUserID
	return r.view, r.err
}

func TestGetForActiveTripUsesAuthenticatedRiderIdentity(t *testing.T) {
	riderUserID := uuid.New()
	expected := View{RideRequestID: uuid.New(), Latitude: 24.8607, Longitude: 67.0011, UpdatedAt: time.Now()}
	repository := &fakeRepository{view: expected}
	service := NewService(repository)

	view, err := service.GetForActiveTrip(context.Background(), riderUserID)
	if err != nil {
		t.Fatalf("GetForActiveTrip returned error: %v", err)
	}
	if repository.riderUserID != riderUserID {
		t.Fatalf("repository rider identity = %s, want %s", repository.riderUserID, riderUserID)
	}
	if view != expected {
		t.Fatalf("view = %#v, want %#v", view, expected)
	}
}
