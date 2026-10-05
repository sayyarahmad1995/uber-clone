package httpapi

import (
	"errors"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
)

func TestRoutePointEndpointPreservesOptionalPlaceID(t *testing.T) {
	latitude := 33.7295
	longitude := 73.0372
	endpoint := (routePoint{
		Latitude:  &latitude,
		Longitude: &longitude,
		PlaceID:   "  poi-place-id  ",
	}).endpoint()

	if endpoint.Point.Latitude != latitude ||
		endpoint.Point.Longitude != longitude ||
		endpoint.PlaceID != "poi-place-id" {
		t.Fatalf("unexpected endpoint: %#v", endpoint)
	}
}

func TestRidePreviewResponseUsesSingleRouteContract(t *testing.T) {
	response := ridePreviewResponse(
		routing.Route{
			DistanceMeters:  12400,
			DurationSeconds: 901,
			EncodedPolyline: "encoded-a",
		},
		pricing.Suggestion{
			Fare: pricing.Fare{
				AmountMinor: 34500,
				Currency:    "PKR",
			},
			PolicyVersion: 4,
		},
	)
	route, ok := response["route"].(map[string]any)
	if !ok {
		t.Fatalf("missing route: %#v", response)
	}
	if route["distance_meters"] != int64(12400) ||
		route["duration_seconds"] != int64(901) ||
		route["encoded_polyline"] != "encoded-a" {
		t.Fatalf("unexpected route response: %#v", route)
	}
	fare, ok := response["suggested_fare"].(map[string]any)
	if !ok ||
		fare["amount_minor"] != int64(34500) ||
		fare["currency"] != "PKR" ||
		response["pricing_policy_version"] != int64(4) {
		t.Fatalf("unexpected pricing response: %#v", response)
	}
	if _, exists := response["routes"]; exists {
		t.Fatalf("alternative routes contract must not be exposed: %#v", response)
	}
}

func TestRoutingErrorStatusContract(t *testing.T) {
	cases := []struct {
		err    error
		status int
	}{
		{routing.ErrInvalidInput, 400},
		{routing.ErrNotFound, 404},
		{routing.ErrUnavailable, 503},
		{routing.ErrProvider, 502},
		{errors.New("unexpected"), 502},
	}
	for _, tc := range cases {
		if got := routingStatus(tc.err); got != tc.status {
			t.Fatalf("error %v: expected %d, got %d", tc.err, tc.status, got)
		}
	}
}
