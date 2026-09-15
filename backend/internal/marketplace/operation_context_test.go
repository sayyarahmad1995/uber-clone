package marketplace

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestOperationContextJSONRoundTripPreservesSnapshotContract(t *testing.T) {
	vehicleID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	want := OperationContext{
		VehicleID:    vehicleID,
		ServiceCode:  "ride_standard",
		ServiceName:  "Standard",
		DriverName:   "Driver One",
		Make:         "Toyota",
		Model:        "Corolla",
		ModelYear:    2022,
		Color:        "White",
		LicensePlate: "ABC-123",
		Fare: OperationFare{
			AmountMinor: 125000,
			Currency:    "PKR",
		},
	}

	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got OperationContext
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got != want {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("decode keys: %v", err)
	}
	wantKeys := []string{
		"vehicle_id",
		"service_code",
		"service_name",
		"driver_name",
		"make",
		"model",
		"model_year",
		"color",
		"license_plate",
		"fare",
	}
	if len(keys) != len(wantKeys) {
		t.Fatalf("JSON key count = %d, want %d; JSON=%s", len(keys), len(wantKeys), raw)
	}
	for _, key := range wantKeys {
		if _, ok := keys[key]; !ok {
			t.Fatalf("JSON missing key %q; JSON=%s", key, raw)
		}
	}

	var fareKeys map[string]json.RawMessage
	if err := json.Unmarshal(keys["fare"], &fareKeys); err != nil {
		t.Fatalf("decode fare keys: %v", err)
	}
	if len(fareKeys) != 2 || fareKeys["amount_minor"] == nil || fareKeys["currency"] == nil {
		t.Fatalf("fare JSON keys = %#v, want amount_minor and currency", fareKeys)
	}
}

func TestOperationContextPointerAcceptsHistoricalJSONNull(t *testing.T) {
	var got *OperationContext
	if err := json.Unmarshal([]byte("null"), &got); err != nil {
		t.Fatalf("Unmarshal(null) error = %v", err)
	}
	if got != nil {
		t.Fatalf("Unmarshal(null) = %#v, want nil", got)
	}
}
