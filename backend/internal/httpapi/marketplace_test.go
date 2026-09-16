package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
)

func TestMarketplaceAssignmentErrorContract(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"ride not open", marketplace.ErrNotOpen, http.StatusConflict, "ride request is not open for offers"},
		{"offer not actionable", marketplace.ErrOfferNotActionable, http.StatusConflict, "ride offer is not actionable"},
		{"driver unavailable", marketplace.ErrDriverUnavailable, http.StatusConflict, "driver is not eligible to offer or assign"},
		{"unexpected error", errors.New("private database details"), http.StatusInternalServerError, "unable to process ride offer"},
	}
	for _, tt := range tests {
		for _, wrapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/wrapped=%t", tt.name, wrapped), func(t *testing.T) {
				err := tt.err
				if wrapped {
					err = fmt.Errorf("select offer: %w", err)
				}
				response := httptest.NewRecorder()
				if !writeMarketplaceAssignmentError(response, err) {
					t.Fatal("expected error response to stop handler processing")
				}
				if response.Code != tt.status {
					t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body.String())
				}
				var body map[string]string
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if len(body) != 1 || body["error"] != tt.message {
					t.Fatalf("unexpected error contract: %s", response.Body.String())
				}
			})
		}
	}
}

func TestMarketplaceAssignmentNilErrorLeavesResponseUntouched(t *testing.T) {
	response := httptest.NewRecorder()
	response.Code = 0
	if writeMarketplaceAssignmentError(response, nil) {
		t.Fatal("nil error must allow handler processing to continue")
	}
	if response.Code != 0 || response.Body.Len() != 0 || len(response.Header()) != 0 {
		t.Fatalf("nil error wrote a response: %#v", response)
	}
}

func TestRiderOfferComparisonContractAndPrivacy(t *testing.T) {
	zero := 0.0
	item := offer.RiderOffer{
		Offer:                offer.Offer{RideRequestID: uuid.New(), DriverUserID: uuid.New(), AmountMinor: 100000, Currency: "PKR", Status: offer.StatusPending},
		Driver:               &offer.DriverSummary{DisplayName: "Sayyar Ahmad"},
		Vehicle:              &offer.VehicleSummary{Make: "Toyota", Model: "Corolla", ModelYear: 2024, Color: "White"},
		Service:              &offer.ServiceSummary{Code: "comfort", DisplayName: "Comfort"},
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
	if len(decoded) != 14 || decoded["matches_proposed_fare"] != true || decoded["selectable"] != true {
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
	service := decoded["service"].(map[string]any)
	if len(service) != 2 || service["code"] != "comfort" || service["display_name"] != "Comfort" {
		t.Fatalf("unexpected service projection: %s", body)
	}
}

func TestLegacyPresentationUsesNullModelYear(t *testing.T) {
	item := offer.RiderOffer{Vehicle: &offer.VehicleSummary{Make: "Toyota", Model: "Corolla", Color: "White"}, Selectable: true}
	body, err := json.Marshal(riderOfferResponse(item))
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response["driver"] != nil || response["service"] != nil || response["vehicle"].(map[string]any)["model_year"] != nil || response["selectable"] != true {
		t.Fatalf("legacy comparison contract: %s", body)
	}
	rr := httptest.NewRecorder()
	writeDriver(rr, 200, driver.Profile{Vehicle: driver.Vehicle{Make: "Toyota", Model: "Corolla"}})
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["vehicle"].(map[string]any)["model_year"] != nil {
		t.Fatalf("legacy Driver contract: %s", rr.Body.String())
	}
}

func TestRiderOfferComparisonKeepsUnavailablePresentationSnapshot(t *testing.T) {
	item := offer.RiderOffer{
		Offer:      offer.Offer{RideRequestID: uuid.New(), DriverUserID: uuid.New(), AmountMinor: 100000, Currency: "PKR", Status: offer.StatusPending},
		Driver:     &offer.DriverSummary{DisplayName: "Driver"},
		Vehicle:    &offer.VehicleSummary{Make: "Toyota", Model: "Corolla", ModelYear: 2024, Color: "White"},
		Service:    &offer.ServiceSummary{Code: "economy", DisplayName: "Economy"},
		Selectable: false,
	}
	body, err := json.Marshal(riderOfferResponse(item))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["pickup_distance_meters"] != nil || decoded["driver"] == nil || decoded["vehicle"] == nil || decoded["service"] == nil {
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
