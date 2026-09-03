package httpapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestRideRequestListResponseUsesStatusProjectionAndEmptyArray(t *testing.T) {
	empty := rideRequestListResponse(nil)
	requests, ok := empty["ride_requests"].([]map[string]any)
	if !ok {
		t.Fatalf("ride_requests type = %T, want []map[string]any", empty["ride_requests"])
	}
	if len(requests) != 0 {
		t.Fatalf("ride_requests length = %d, want 0", len(requests))
	}

	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	assignedAt := createdAt.Add(time.Minute)
	startedAt := assignedAt.Add(time.Minute)
	request := ride.Request{
		ID:           uuid.New(),
		Pickup:       ride.Location{Latitude: 24.86, Longitude: 67.01},
		Destination:  ride.Location{Latitude: 24.91, Longitude: 67.08},
		BookingMode:  ride.BookingModeOffers,
		ProposedFare: &ride.Money{AmountMinor: 100000, Currency: "PKR"},
		Status:       ride.StatusRequested,
		CreatedAt:    createdAt,
	}
	assignedTrip := trip.Trip{
		RideRequestID: request.ID,
		DriverUserID:  uuid.New(),
		Status:        trip.StatusInProgress,
		AssignedAt:    assignedAt,
		StartedAt:     &startedAt,
	}

	response := rideRequestListResponse([]ridestatus.View{{RideRequest: request, Trip: &assignedTrip}})
	requests, ok = response["ride_requests"].([]map[string]any)
	if !ok || len(requests) != 1 {
		t.Fatalf("unexpected ride_requests payload: %#v", response["ride_requests"])
	}
	item := requests[0]
	if item["id"] != request.ID || item["proposed_fare"] == nil {
		t.Fatalf("unexpected ride request projection: %#v", item)
	}
	if _, exists := item["booking_mode"]; exists {
		t.Fatal("list response must not expose legacy booking_mode")
	}
	tripPayload, ok := item["trip"].(map[string]any)
	if !ok || tripPayload["status"] != trip.StatusInProgress || tripPayload["driver_user_id"] != assignedTrip.DriverUserID {
		t.Fatalf("unexpected trip projection: %#v", item["trip"])
	}
	if _, exists := item["rider_user_id"]; exists {
		t.Fatal("list response must not expose rider_user_id")
	}
}
