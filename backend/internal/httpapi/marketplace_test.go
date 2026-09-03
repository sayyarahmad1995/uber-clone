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
		Driver:               &offer.DriverSummary{DisplayName: "Sayyar Ahmad"},
		Vehicle:              &offer.VehicleSummary{Make: "Toyota", Model: "Corolla", ModelYear: 2024, Color: "White"},
		PickupDistanceMeters: &zero,
		MatchesProposedFare:  true,
		Selectable:           true,
	}
	body, err := json.Marshal(riderOfferResponse(item))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 12 || decoded["matches_proposed_fare"] != true || decoded["selectable"] != true {
		t.Fatalf("unexpected comparison contract: %s", body)
	}
	for _, key := range []string{"latitude", "longitude", "driver_location", "license_plate", "email", "phone", "rider_user_id", "eta"} {
		if _, present := decoded[key]; present {
			t.Fatalf("unexpected pre-assignment field %s", key)
		}
	}
	if decoded["pickup_distance_meters"] != float64(0) {
		t.Fatalf("zero distance was lost: %s", body)
	}
	driver := decoded["driver"].(map[string]any)
	if len(driver) != 1 || driver["display_name"] != "Sayyar Ahmad" {
		t.Fatalf("unexpected driver projection: %s", body)
	}
	vehicle := decoded["vehicle"].(map[string]any)
	if len(vehicle) != 4 || vehicle["make"] != "Toyota" || vehicle["model"] != "Corolla" || vehicle["model_year"] != float64(2024) || vehicle["color"] != "White" {
		t.Fatalf("unexpected vehicle projection: %s", body)
	}
}

func TestRiderOfferComparisonKeepsUnavailablePresentationSnapshot(t *testing.T) {
	item := offer.RiderOffer{
		Offer:       offer.Offer{RideRequestID: uuid.New(), DriverUserID: uuid.New(), AmountMinor: 100000, Currency: "PKR", Status: offer.StatusPending},
		Driver:      &offer.DriverSummary{DisplayName: "Driver"},
		Vehicle:     &offer.VehicleSummary{Make: "Toyota", Model: "Corolla", ModelYear: 2024, Color: "White"},
		Selectable:  false,
	}
	body, err := json.Marshal(riderOfferResponse(item))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["pickup_distance_meters"] != nil || decoded["driver"] == nil || decoded["vehicle"] == nil {
		t.Fatalf("unavailable offer should retain known public presentation but not stale distance: %s", body)
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
