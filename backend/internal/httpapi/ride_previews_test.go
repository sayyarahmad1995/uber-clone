package httpapi

import (
	"errors"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
)

func TestRidePreviewResponseUsesNormalizedRouteOptionsContract(t *testing.T) {
	response := ridePreviewResponse(routing.Preview{Routes: []routing.Route{
		{
			ID:              "route-0",
			Recommended:     true,
			DistanceMeters:  12400,
			DurationSeconds: 901,
			EncodedPolyline: "encoded-a",
		},
		{
			ID:              "route-1",
			DistanceMeters:  11800,
			DurationSeconds: 960,
			EncodedPolyline: "encoded-b",
		},
	}})
	routes, ok := response["routes"].([]map[string]any)
	if !ok || len(routes) != 2 {
		t.Fatalf("missing route options: %#v", response)
	}
	if routes[0]["id"] != "route-0" ||
		routes[0]["recommended"] != true ||
		routes[0]["distance_meters"] != int64(12400) ||
		routes[0]["duration_seconds"] != int64(901) ||
		routes[0]["encoded_polyline"] != "encoded-a" {
		t.Fatalf("unexpected recommended route response: %#v", routes[0])
	}
	if routes[1]["id"] != "route-1" || routes[1]["recommended"] != false {
		t.Fatalf("unexpected alternate route response: %#v", routes[1])
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
