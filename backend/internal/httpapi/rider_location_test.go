package httpapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/riderlocation"
)

func TestRiderActiveTripDriverLocationResponseDoesNotExposeDriverIdentity(t *testing.T) {
	view := riderlocation.View{
		RideRequestID: uuid.New(),
		Latitude:      24.8607,
		Longitude:     67.0011,
		UpdatedAt:     time.Now(),
	}

	response := riderActiveTripDriverLocationResponse(view)
	if response["ride_request_id"] != view.RideRequestID || response["latitude"] != view.Latitude || response["longitude"] != view.Longitude || response["updated_at"] != view.UpdatedAt {
		t.Fatalf("unexpected response: %#v", response)
	}
	if _, exists := response["driver_user_id"]; exists {
		t.Fatal("response must not expose driver_user_id")
	}
	if _, exists := response["rider_user_id"]; exists {
		t.Fatal("response need not echo authenticated rider_user_id")
	}
}
