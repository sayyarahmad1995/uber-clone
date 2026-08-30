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

	pickup, destination, ok := body.locations()
	if !ok {
		t.Fatal("locations reported complete request as incomplete")
	}
	if pickup.Latitude != 24.8607 || pickup.Longitude != 67.0011 {
		t.Fatalf("unexpected pickup: %#v", pickup)
	}
	if destination.Latitude != 24.9056 || destination.Longitude != 67.0822 {
		t.Fatalf("unexpected destination: %#v", destination)
	}
}

func TestCreateRideRequestBodyRequiresCompleteLocations(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing pickup", body: `{"destination":{"latitude":24.9056,"longitude":67.0822}}`},
		{name: "missing destination", body: `{"pickup":{"latitude":24.8607,"longitude":67.0011}}`},
		{name: "missing pickup latitude", body: `{"pickup":{"longitude":67.0011},"destination":{"latitude":24.9056,"longitude":67.0822}}`},
		{name: "missing pickup longitude", body: `{"pickup":{"latitude":24.8607},"destination":{"latitude":24.9056,"longitude":67.0822}}`},
		{name: "missing destination latitude", body: `{"pickup":{"latitude":24.8607,"longitude":67.0011},"destination":{"longitude":67.0822}}`},
		{name: "missing destination longitude", body: `{"pickup":{"latitude":24.8607,"longitude":67.0011},"destination":{"latitude":24.9056}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body createRideRequestBody
			if err := json.NewDecoder(strings.NewReader(tt.body)).Decode(&body); err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}
			if _, _, ok := body.locations(); ok {
				t.Fatal("locations accepted incomplete request")
			}
		})
	}
}

func TestCreateRideRequestBodyAllowsPresentZeroCoordinates(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{
		"pickup":{"latitude":0,"longitude":0},
		"destination":{"latitude":0,"longitude":0}
	}`)).Decode(&body); err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	pickup, destination, ok := body.locations()
	if !ok {
		t.Fatal("locations rejected present zero coordinates")
	}
	if pickup.Latitude != 0 || pickup.Longitude != 0 || destination.Latitude != 0 || destination.Longitude != 0 {
		t.Fatalf("unexpected zero-coordinate mapping: pickup=%#v destination=%#v", pickup, destination)
	}
}
