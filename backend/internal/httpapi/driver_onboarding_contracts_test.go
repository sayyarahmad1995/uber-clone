package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDriverOnboardingRequestIncludesSelectedServiceAndVehicle(t *testing.T) {
	var request driverOnboardingRequest
	if err := json.NewDecoder(strings.NewReader(`{
		"display_name":"Test Driver",
		"service_code":"comfort",
		"vehicle":{
			"make":"Toyota",
			"model":"Corolla",
			"model_year":2024,
			"color":"White",
			"license_plate":"ABC-123"
		}
	}`)).Decode(&request); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	input := request.input()
	if input.DisplayName != "Test Driver" || input.ServiceCode != "comfort" {
		t.Fatalf("unexpected Driver/service input: %#v", input)
	}
	if input.Vehicle.Make != "Toyota" || input.Vehicle.Model != "Corolla" || input.Vehicle.ModelYear != 2024 || input.Vehicle.Color != "White" || input.Vehicle.LicensePlate != "ABC-123" {
		t.Fatalf("unexpected vehicle input: %#v", input.Vehicle)
	}
}
