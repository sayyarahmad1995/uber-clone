package httpapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestDriverTripHistoryResponseUsesEmptyArrayAndDoesNotExposeIdentity(t *testing.T) {
	empty := driverTripHistoryResponse(nil)
	trips, ok := empty["trips"].([]map[string]any)
	if !ok {
		t.Fatalf("trips type = %T, want []map[string]any", empty["trips"])
	}
	if len(trips) != 0 {
		t.Fatalf("trips length = %d, want 0", len(trips))
	}

	completedAt := time.Now().UTC()
	view := drivertrip.View{
		RideRequestID: uuid.New(),
		Pickup:        ride.Location{Latitude: 24.8607, Longitude: 67.0011},
		Destination:   ride.Location{Latitude: 24.9056, Longitude: 67.0822},
		Status:        trip.StatusCompleted,
		AssignedAt:    completedAt.Add(-10 * time.Minute),
		StartedAt:     ptrTime(completedAt.Add(-8 * time.Minute)),
		CompletedAt:   &completedAt,
	}
	response := driverTripHistoryResponse([]drivertrip.View{view})
	trips = response["trips"].([]map[string]any)
	if len(trips) != 1 {
		t.Fatalf("trips length = %d, want 1", len(trips))
	}
	item := trips[0]
	if item["ride_request_id"] != view.RideRequestID {
		t.Fatalf("ride_request_id = %v, want %v", item["ride_request_id"], view.RideRequestID)
	}
	if item["status"] != trip.StatusCompleted {
		t.Fatalf("status = %v, want %v", item["status"], trip.StatusCompleted)
	}
	if _, exists := item["rider_user_id"]; exists {
		t.Fatal("history item must not expose rider_user_id")
	}
	if _, exists := item["driver_user_id"]; exists {
		t.Fatal("history item must not redundantly expose driver_user_id")
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
