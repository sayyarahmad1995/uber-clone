package httpapi

import (
    "encoding/json"
    "testing"

    "github.com/sayyarahmad1995/uber-clone/backend/internal/rideservice"
)

func TestRideServicesResponseIsRiderPresentationOnly(t *testing.T) {
    response := rideServicesResponse([]rideservice.Option{
        {Code: "economy", DisplayName: "Economy", Description: "Standard ride", DisplayOrder: 10, PresentationToken: "car"},
        {Code: "test_third", DisplayName: "Another car", Description: "", DisplayOrder: 20, PresentationToken: "future-token"},
    })
    encoded, err := json.Marshal(response)
    if err != nil {
        t.Fatal(err)
    }
    var decoded struct {
        Services []map[string]any `json:"services"`
    }
    if err := json.Unmarshal(encoded, &decoded); err != nil {
        t.Fatal(err)
    }
    if len(decoded.Services) != 2 {
        t.Fatalf("expected two services, got %d", len(decoded.Services))
    }
    if decoded.Services[1]["code"] != "test_third" || decoded.Services[1]["presentation_token"] != "future-token" {
        t.Fatalf("unknown service presentation lost: %#v", decoded.Services[1])
    }
    for _, s := range decoded.Services {
        if _, ok := s["minimum_model_year"]; ok {
            t.Fatal("Rider response leaked Driver-only vehicle requirements")
        }
        if _, ok := s["implied_service_code"]; ok {
            t.Fatal("Rider response leaked Driver-only eligibility")
        }
    }
}

func TestRideServicesResponseEmptyIsAnArray(t *testing.T) {
    encoded, err := json.Marshal(rideServicesResponse(nil))
    if err != nil {
        t.Fatal(err)
    }
    if string(encoded) != `{"services":[]}` {
        t.Fatalf("unexpected empty Rider catalog: %s", encoded)
    }
}
