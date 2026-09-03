package httpapi

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
)

func TestRiderOfferComparisonContractAndPrivacy(t *testing.T) {
	zero := 0.0
	item := offer.RiderOffer{
		Offer:                offer.Offer{RideRequestID: uuid.New(), DriverUserID: uuid.New(), AmountMinor: 100000, Currency: "PKR", Status: offer.StatusPending},
		Vehicle:              &offer.VehicleSummary{Make: "Toyota", Model: "Corolla", Color: "White"},
		PickupDistanceMeters: &zero, MatchesProposedFare: true, Selectable: true,
	}
	for _, unavailable := range []bool{false, true} {
		if unavailable {
			item.PickupDistanceMeters = nil
			item.Selectable = false
			item.Vehicle = nil
		}
		body, err := json.Marshal(riderOfferResponse(item))
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatal(err)
		}
		if len(decoded) != 11 || decoded["matches_proposed_fare"] != true || decoded["selectable"] != !unavailable {
			t.Fatalf("unexpected comparison contract: %s", body)
		}
		for _, key := range []string{"latitude", "longitude", "driver_location", "license_plate", "email", "phone", "rider_user_id", "eta"} {
			if _, present := decoded[key]; present {
				t.Fatalf("unexpected pre-assignment field %s", key)
			}
		}
		if unavailable {
			if decoded["pickup_distance_meters"] != nil || decoded["vehicle"] != nil {
				t.Fatalf("unknown values must be null: %s", body)
			}
		} else {
			if decoded["pickup_distance_meters"] != float64(0) {
				t.Fatalf("zero distance was lost: %s", body)
			}
			vehicle := decoded["vehicle"].(map[string]any)
			if len(vehicle) != 3 || vehicle["make"] != "Toyota" || vehicle["model"] != "Corolla" || vehicle["color"] != "White" {
				t.Fatalf("unexpected vehicle projection: %s", body)
			}
		}
	}
}

func TestDiscoveryDistanceContract(t *testing.T) {
	body, err := json.Marshal(driverMarketplaceItemResponse(offer.DiscoveryItem{PickupDistanceMeters: 123.5}))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["pickup_distance_meters"] != 123.5 {
		t.Fatalf("missing distance: %s", body)
	}
	if _, present := decoded["rider_user_id"]; present {
		t.Fatal("Rider identity exposed")
	}
}
