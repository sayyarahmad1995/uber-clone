package trip

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestOperationContextJSONRoundTripPreservesCompleteSnapshot(t *testing.T) {
	original := []byte(`{
		"driver_name":"Ayesha Khan",
		"vehicle_id":"11111111-1111-1111-1111-111111111111",
		"make":"Toyota",
		"model":"Corolla",
		"model_year":2024,
		"color":"White",
		"license_plate":"XYZ 987",
		"service_code":"comfort",
		"service_name":"Comfort",
		"fare":{"amount_minor":125000,"currency":"PKR"}
	}`)

	var context OperationContext
	if err := json.Unmarshal(original, &context); err != nil {
		t.Fatal(err)
	}
	roundTripped, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}

	var want, got map[string]any
	if err := json.Unmarshal(original, &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(roundTripped, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("operation context changed during JSON round trip:\n got: %s\nwant: %s", roundTripped, original)
	}
}
