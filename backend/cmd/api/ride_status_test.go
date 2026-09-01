package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestRideRequestStatusResponseWithoutTrip(t *testing.T) {
	request := ride.Request{
		ID:          uuid.New(),
		Pickup:      ride.Location{Latitude: 24.86, Longitude: 67.01},
		Destination: ride.Location{Latitude: 24.91, Longitude: 67.08},
		BookingMode: ride.BookingModeOffers,
		ProposedFare: &ride.Money{
			AmountMinor: 100000,
			Currency:    "PKR",
		},
		Status:    ride.StatusRequested,
		CreatedAt: time.Now(),
	}

	response := rideRequestStatusResponse(request, nil)
	if response["trip"] != nil {
		t.Fatalf("trip = %#v, want nil", response["trip"])
	}
	if response["id"] != request.ID {
		t.Fatalf("id = %v, want %s", response["id"], request.ID)
	}
}

func TestRideRequestStatusResponseIncludesTripExecutionState(t *testing.T) {
	request := ride.Request{
		ID:          uuid.New(),
		Pickup:      ride.Location{Latitude: 24.86, Longitude: 67.01},
		Destination: ride.Location{Latitude: 24.91, Longitude: 67.08},
		BookingMode: ride.BookingModeAutomatic,
		Status:      ride.StatusRequested,
		CreatedAt:   time.Now(),
	}
	assignedAt := time.Now().Add(-time.Minute)
	startedAt := time.Now()
	assignedTrip := trip.Trip{
		RideRequestID: request.ID,
		DriverUserID:  uuid.New(),
		Status:        trip.StatusInProgress,
		AssignedAt:    assignedAt,
		StartedAt:     &startedAt,
	}

	response := rideRequestStatusResponse(request, &assignedTrip)
	tripResponse, ok := response["trip"].(map[string]any)
	if !ok {
		t.Fatalf("trip = %#v, want response object", response["trip"])
	}
	if tripResponse["status"] != trip.StatusInProgress {
		t.Fatalf("trip status = %v, want %s", tripResponse["status"], trip.StatusInProgress)
	}
	if tripResponse["driver_user_id"] != assignedTrip.DriverUserID {
		t.Fatalf("driver_user_id = %v, want %s", tripResponse["driver_user_id"], assignedTrip.DriverUserID)
	}
}
