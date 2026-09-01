package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestDriverCurrentTripResponseProjectsExecutionStateWithoutIdentityLeakage(t *testing.T) {
	assignedAt := time.Now().Add(-time.Minute)
	startedAt := time.Now()
	view := drivertrip.View{
		RideRequestID: uuid.New(),
		Pickup:        ride.Location{Latitude: 24.86, Longitude: 67.01},
		Destination:   ride.Location{Latitude: 24.91, Longitude: 67.08},
		Status:        trip.StatusInProgress,
		AssignedAt:    assignedAt,
		StartedAt:     &startedAt,
	}

	response := driverCurrentTripResponse(view)
	if response["ride_request_id"] != view.RideRequestID {
		t.Fatalf("ride_request_id = %v, want %s", response["ride_request_id"], view.RideRequestID)
	}
	if response["status"] != trip.StatusInProgress {
		t.Fatalf("status = %v, want %s", response["status"], trip.StatusInProgress)
	}
	if response["assigned_at"] != assignedAt {
		t.Fatalf("assigned_at = %v, want %v", response["assigned_at"], assignedAt)
	}
	if response["started_at"] != &startedAt {
		t.Fatalf("started_at = %v, want %v", response["started_at"], &startedAt)
	}
	if _, exists := response["rider_user_id"]; exists {
		t.Fatal("response must not expose rider_user_id")
	}
	if _, exists := response["driver_user_id"]; exists {
		t.Fatal("response need not echo authenticated driver_user_id")
	}
}
