package matching

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	result Result
	err    error
	rideID uuid.UUID
	rider  uuid.UUID
}

func (f *fakeRepository) Match(_ context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	f.rideID = rideRequestID
	f.rider = riderUserID
	return f.result, f.err
}

func TestServiceMatchDelegatesApplicationOwnedIDs(t *testing.T) {
	rideID := uuid.New()
	riderID := uuid.New()
	driverID := uuid.New()
	createdAt := time.Now().UTC()
	repository := &fakeRepository{result: Result{
		Candidate: Candidate{RideRequestID: rideID, DriverUserID: driverID, CreatedAt: createdAt},
		Created:   true,
	}}

	result, err := NewService(repository).Match(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if repository.rideID != rideID || repository.rider != riderID {
		t.Fatalf("unexpected delegated ids: ride=%s rider=%s", repository.rideID, repository.rider)
	}
	if !result.Created || result.Candidate.DriverUserID != driverID {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestServiceMatchPreservesRepositoryBusinessError(t *testing.T) {
	repository := &fakeRepository{err: ErrNoEligibleDriver}
	_, err := NewService(repository).Match(context.Background(), uuid.New(), uuid.New())
	if err != ErrNoEligibleDriver {
		t.Fatalf("expected ErrNoEligibleDriver, got %v", err)
	}
}
