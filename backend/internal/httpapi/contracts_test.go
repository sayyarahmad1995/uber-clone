package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/matching"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestOnboardDriverRequestUsesNestedSnakeCaseVehicleContract(t *testing.T) {
	var request onboardDriverRequest
	if err := json.NewDecoder(strings.NewReader(`{"vehicle":{"make":"Toyota","model":"Corolla","color":"White","license_plate":"ABC-123"}}`)).Decode(&request); err != nil {
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

func TestWriteCandidateDecisionUsesApplicationOwnedContract(t *testing.T) {
	decidedAt := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	candidate := matching.Candidate{RideRequestID: uuid.New(), DriverUserID: uuid.New(), Status: matching.CandidateStatusAccepted, CreatedAt: decidedAt.Add(-time.Minute), DecidedAt: &decidedAt}
	recorder := httptest.NewRecorder()
	writeCandidateDecision(recorder, candidate)
	if recorder.Code != 200 {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var payload struct {
		RideRequestID uuid.UUID                `json:"ride_request_id"`
		DriverUserID  uuid.UUID                `json:"driver_user_id"`
		Status        matching.CandidateStatus `json:"status"`
		CreatedAt     time.Time                `json:"created_at"`
		DecidedAt     *time.Time               `json:"decided_at"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.RideRequestID != candidate.RideRequestID || payload.DriverUserID != candidate.DriverUserID || payload.Status != matching.CandidateStatusAccepted || payload.DecidedAt == nil || !payload.DecidedAt.Equal(decidedAt) {
		t.Fatalf("unexpected decision payload: %#v", payload)
	}
}

func TestCreateRideRequestBodyUsesUnifiedMarketplaceContract(t *testing.T) {
	var body createRideRequestBody
	err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":24.8607,"longitude":67.0011},"destination":{"latitude":24.9056,"longitude":67.0822},"proposed_fare":{"amount_minor":70000,"currency":"PKR"}}`)).Decode(&body)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	input, ok := body.input()
	if !ok || input.ProposedFare == nil || input.ProposedFare.AmountMinor != 70000 || input.ProposedFare.Currency != "PKR" {
		t.Fatalf("unexpected input: %#v, complete=%v", input, ok)
	}
}

func TestCreateRideRequestBodyRequiresCompleteLocations(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":24.8607},"destination":{"latitude":24.9056,"longitude":67.0822},"proposed_fare":{"amount_minor":70000,"currency":"PKR"}}`)).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body.input(); ok {
		t.Fatal("input accepted incomplete locations")
	}
}

func TestCreateRideRequestBodyRequiresProposedFare(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":0,"longitude":0},"destination":{"latitude":0,"longitude":0}}`)).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body.input(); ok {
		t.Fatal("input accepted missing proposed fare")
	}
}

func TestCreateRideRequestBodyRequiresCompleteFare(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":0,"longitude":0},"destination":{"latitude":0,"longitude":0},"proposed_fare":{"currency":"PKR"}}`)).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body.input(); ok {
		t.Fatal("input accepted incomplete fare")
	}
}

func TestCreateRideRequestBodyAllowsPresentZeroCoordinates(t *testing.T) {
	var body createRideRequestBody
	if err := json.NewDecoder(strings.NewReader(`{"pickup":{"latitude":0,"longitude":0},"destination":{"latitude":0,"longitude":0},"proposed_fare":{"amount_minor":1000,"currency":"PKR"}}`)).Decode(&body); err != nil {
		t.Fatal(err)
	}
	input, ok := body.input()
	if !ok || input.Pickup.Latitude != 0 || input.Pickup.Longitude != 0 || input.Destination.Latitude != 0 || input.Destination.Longitude != 0 {
		t.Fatalf("unexpected input: %#v, complete=%v", input, ok)
	}
}

func TestRideRequestStatusResponseWithoutTrip(t *testing.T) {
	request := ride.Request{ID: uuid.New(), Pickup: ride.Location{Latitude: 24.86, Longitude: 67.01}, Destination: ride.Location{Latitude: 24.91, Longitude: 67.08}, BookingMode: ride.BookingModeOffers, ProposedFare: &ride.Money{AmountMinor: 100000, Currency: "PKR"}, Status: ride.StatusRequested, CreatedAt: time.Now()}
	response := rideRequestStatusResponse(request, nil)
	if response["trip"] != nil || response["id"] != request.ID {
		t.Fatalf("unexpected response: %#v", response)
	}
	if _, exists := response["booking_mode"]; exists {
		t.Fatal("unified Rider response must not expose booking_mode")
	}
}

func TestRideRequestStatusResponseIncludesTripExecutionState(t *testing.T) {
	request := ride.Request{ID: uuid.New(), Pickup: ride.Location{Latitude: 24.86, Longitude: 67.01}, Destination: ride.Location{Latitude: 24.91, Longitude: 67.08}, BookingMode: ride.BookingModeAutomatic, Status: ride.StatusRequested, CreatedAt: time.Now()}
	assignedAt := time.Now().Add(-time.Minute)
	startedAt := time.Now()
	assignedTrip := trip.Trip{RideRequestID: request.ID, DriverUserID: uuid.New(), Status: trip.StatusInProgress, AssignedAt: assignedAt, StartedAt: &startedAt}
	response := rideRequestStatusResponse(request, &assignedTrip)
	tripPayload, ok := response["trip"].(map[string]any)
	if !ok || tripPayload["status"] != trip.StatusInProgress || tripPayload["driver_user_id"] != assignedTrip.DriverUserID {
		t.Fatalf("unexpected trip response: %#v", response["trip"])
	}
}

func TestDriverCurrentTripResponseProjectsExecutionStateWithoutIdentityLeakage(t *testing.T) {
	assignedAt := time.Now().Add(-time.Minute)
	startedAt := time.Now()
	view := drivertrip.View{RideRequestID: uuid.New(), Pickup: ride.Location{Latitude: 24.86, Longitude: 67.01}, Destination: ride.Location{Latitude: 24.91, Longitude: 67.08}, Status: trip.StatusInProgress, AssignedAt: assignedAt, StartedAt: &startedAt}
	response := driverCurrentTripResponse(view)
	if response["ride_request_id"] != view.RideRequestID || response["status"] != trip.StatusInProgress || response["assigned_at"] != assignedAt || response["started_at"] != &startedAt {
		t.Fatalf("unexpected current trip response: %#v", response)
	}
	if _, exists := response["rider_user_id"]; exists {
		t.Fatal("response must not expose rider_user_id")
	}
	if _, exists := response["driver_user_id"]; exists {
		t.Fatal("response need not echo authenticated driver_user_id")
	}
}
