package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreateRideRequestBodyUsesApplicationOwnedLocationContract(t *testing.T) {
	var body createRideRequestBody
	err := json.NewDecoder(strings.NewReader(`{
		"pickup":{"latitude":24.8607,"longitude":67.0011},
		"destination":{"latitude":24.9056,"longitude":67.0822}
	}`)).Decode(&body)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if body.Pickup.Latitude != 24.8607 || body.Pickup.Longitude != 67.0011 {
		t.Fatalf("unexpected pickup: %#v", body.Pickup)
	}
	if body.Destination.Latitude != 24.9056 || body.Destination.Longitude != 67.0822 {
		t.Fatalf("unexpected destination: %#v", body.Destination)
	}
}
