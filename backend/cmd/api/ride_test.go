package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
)

func TestCreateRideRequestBodyUsesApplicationOwnedContract(t *testing.T) {
	var body createRideRequestBody
	err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":24.8607,"longitude":67.0011},"destination":{"latitude":24.9056,"longitude":67.0822},"booking_mode":"offers","proposed_fare":{"amount_minor":70000,"currency":"PKR"}}`)).Decode(&body)
	if err != nil { t.Fatalf("Decode returned error: %v",err) }
	input,ok := body.input(); if !ok { t.Fatal("input reported complete request as incomplete") }
	if input.BookingMode != ride.BookingModeOffers { t.Fatalf("unexpected booking mode: %q",input.BookingMode) }
	if input.ProposedFare == nil || input.ProposedFare.AmountMinor != 70000 || input.ProposedFare.Currency != "PKR" { t.Fatalf("unexpected fare: %#v",input.ProposedFare) }
}

func TestCreateRideRequestBodyRequiresCompleteLocations(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":24.8607},"destination":{"latitude":24.9056,"longitude":67.0822}}`)).Decode(&body); err != nil { t.Fatal(err) }
	if _,ok := body.input(); ok { t.Fatal("input accepted incomplete locations") }
}

func TestCreateRideRequestBodyRequiresCompleteFareWhenPresent(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":0,"longitude":0},"destination":{"latitude":0,"longitude":0},"booking_mode":"offers","proposed_fare":{"currency":"PKR"}}`)).Decode(&body); err != nil { t.Fatal(err) }
	if _,ok := body.input(); ok { t.Fatal("input accepted incomplete fare") }
}

func TestCreateRideRequestBodyAllowsPresentZeroCoordinates(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":0,"longitude":0},"destination":{"latitude":0,"longitude":0}}`)).Decode(&body); err != nil { t.Fatal(err) }
	input,ok := body.input(); if !ok { t.Fatal("input rejected present zero coordinates") }
	if input.Pickup.Latitude != 0 || input.Pickup.Longitude != 0 || input.Destination.Latitude != 0 || input.Destination.Longitude != 0 { t.Fatalf("unexpected coordinates: %#v",input) }
}
