package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOnboardDriverRequestUsesNestedSnakeCaseVehicleContract(t *testing.T) {
	var request onboardDriverRequest
	if err := json.NewDecoder(strings.NewReader(`{
		"vehicle": {
			"make": "Toyota",
			"model": "Corolla",
			"color": "White",
			"license_plate": "ABC-123"
		}
	}`)).Decode(&request); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	vehicle := request.vehicleInput()
	if vehicle.Make != "Toyota" || vehicle.Model != "Corolla" || vehicle.Color != "White" || vehicle.LicensePlate != "ABC-123" {
		t.Fatalf("unexpected vehicle: %#v", vehicle)
	}
}

func TestDriverAvailabilityRequestUsesIsOnlineContract(t *testing.T) {
	var request driverAvailabilityRequest
	if err := json.NewDecoder(strings.NewReader(`{"is_online":true}`)).Decode(&request); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if !request.IsOnline {
		t.Fatal("expected is_online=true")
	}
}
