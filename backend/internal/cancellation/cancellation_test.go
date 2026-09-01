package cancellation

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	method        string
	rideRequestID uuid.UUID
	userID        uuid.UUID
	result        Result
	err           error
}

func (r *fakeRepository) CancelByRider(_ context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	r.method = "rider"
	r.rideRequestID = rideRequestID
	r.userID = riderUserID
	return r.result, r.err
}

func (r *fakeRepository) CancelByDriver(_ context.Context, rideRequestID, driverUserID uuid.UUID) (Result, error) {
	r.method = "driver"
	r.rideRequestID = rideRequestID
	r.userID = driverUserID
	return r.result, r.err
}

func TestCancelByRiderPreservesOwnershipScope(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	rideRequestID := uuid.New()
	riderUserID := uuid.New()

	if _, err := service.CancelByRider(context.Background(), rideRequestID, riderUserID); err != nil {
		t.Fatalf("CancelByRider returned error: %v", err)
	}
	if repository.method != "rider" || repository.rideRequestID != rideRequestID || repository.userID != riderUserID {
		t.Fatalf("repository call = %s %s %s", repository.method, repository.rideRequestID, repository.userID)
	}
}

func TestCancelByDriverPreservesAssignmentScope(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	rideRequestID := uuid.New()
	driverUserID := uuid.New()

	if _, err := service.CancelByDriver(context.Background(), rideRequestID, driverUserID); err != nil {
		t.Fatalf("CancelByDriver returned error: %v", err)
	}
	if repository.method != "driver" || repository.rideRequestID != rideRequestID || repository.userID != driverUserID {
		t.Fatalf("repository call = %s %s %s", repository.method, repository.rideRequestID, repository.userID)
	}
}
