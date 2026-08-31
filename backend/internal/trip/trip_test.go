package trip

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	acceptance Acceptance
	trip       Trip
	err        error
	rideID     uuid.UUID
	driverID   uuid.UUID
	operation  string
}

func (f *fakeRepository) Accept(_ context.Context, rideRequestID, driverUserID uuid.UUID) (Acceptance, error) {
	f.rideID, f.driverID, f.operation = rideRequestID, driverUserID, "accept"
	return f.acceptance, f.err
}

func (f *fakeRepository) Start(_ context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	f.rideID, f.driverID, f.operation = rideRequestID, driverUserID, "start"
	return f.trip, f.err
}

func (f *fakeRepository) Complete(_ context.Context, rideRequestID, driverUserID uuid.UUID) (Trip, error) {
	f.rideID, f.driverID, f.operation = rideRequestID, driverUserID, "complete"
	return f.trip, f.err
}

func TestServiceDelegatesDriverOwnedTripLifecycle(t *testing.T) {
	rideID := uuid.New()
	driverID := uuid.New()
	repository := &fakeRepository{trip: Trip{RideRequestID: rideID, DriverUserID: driverID}}
	service := NewService(repository)

	if _, err := service.Start(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("start: %v", err)
	}
	if repository.operation != "start" || repository.rideID != rideID || repository.driverID != driverID {
		t.Fatalf("unexpected start delegation: %#v", repository)
	}

	if _, err := service.Complete(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if repository.operation != "complete" || repository.rideID != rideID || repository.driverID != driverID {
		t.Fatalf("unexpected complete delegation: %#v", repository)
	}
}

func TestServiceAcceptancePreservesRepositoryBusinessError(t *testing.T) {
	repository := &fakeRepository{err: ErrAssignmentResolved}
	_, err := NewService(repository).Accept(context.Background(), uuid.New(), uuid.New())
	if err != ErrAssignmentResolved {
		t.Fatalf("expected ErrAssignmentResolved, got %v", err)
	}
}
