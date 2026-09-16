package marketplace_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type fakeAssignmentRepository struct {
	result trip.Trip
	err    error

	rideRequestID   uuid.UUID
	riderUserID     uuid.UUID
	driverUserID    uuid.UUID
	expectedVersion time.Time
}

func (f *fakeAssignmentRepository) SelectOffer(
	_ context.Context,
	rideRequestID uuid.UUID,
	riderUserID uuid.UUID,
	driverUserID uuid.UUID,
	expectedVersion time.Time,
) (trip.Trip, error) {
	f.rideRequestID = rideRequestID
	f.riderUserID = riderUserID
	f.driverUserID = driverUserID
	f.expectedVersion = expectedVersion
	return f.result, f.err
}

func TestAssignmentServiceOwnsRiderOfferSelection(t *testing.T) {
	rideID := uuid.New()
	riderID := uuid.New()
	driverID := uuid.New()
	version := time.Now().UTC()

	repository := &fakeAssignmentRepository{
		result: trip.Trip{
			RideRequestID: rideID,
			RiderUserID:   riderID,
			DriverUserID:  driverID,
		},
	}

	var service marketplace.AssignmentService = marketplace.NewAssignmentService(repository)

	got, err := service.SelectOffer(
		context.Background(),
		rideID,
		riderID,
		driverID,
		version,
	)
	if err != nil {
		t.Fatalf("select offer: %v", err)
	}

	if got.RideRequestID != rideID ||
		got.RiderUserID != riderID ||
		got.DriverUserID != driverID {
		t.Fatalf("unexpected Trip: %#v", got)
	}

	if repository.rideRequestID != rideID ||
		repository.riderUserID != riderID ||
		repository.driverUserID != driverID ||
		!repository.expectedVersion.Equal(version) {
		t.Fatalf("unexpected assignment delegation: %#v", repository)
	}
}
