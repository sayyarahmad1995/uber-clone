package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
)

func TestPreviewRequestsTrafficAwareAlternativeRoutesAndNormalizesOptions(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/directions/v2:computeRoutes" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Goog-Api-Key") != "server-key" {
			t.Fatal("server API key header missing")
		}
		if got := r.Header.Get("X-Goog-FieldMask"); got != fieldMask || strings.Contains(got, "*") {
			t.Fatalf("unexpected field mask: %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":12345,"duration":"901.4s","routeLabels":["DEFAULT_ROUTE"],"polyline":{"encodedPolyline":"_p~iF~ps|U_ulLnnqC_mqNvxq"}},{"distanceMeters":11900,"duration":"960s","routeLabels":["DEFAULT_ROUTE_ALTERNATE"],"polyline":{"encodedPolyline":"_p~iF~ps|U????"}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	preview, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if body["travelMode"] != "DRIVE" ||
		body["routingPreference"] != "TRAFFIC_AWARE" ||
		body["computeAlternativeRoutes"] != true {
		t.Fatalf("unexpected route options: %#v", body)
	}
	if body["polylineQuality"] != "OVERVIEW" || body["polylineEncoding"] != "ENCODED_POLYLINE" {
		t.Fatalf("unexpected polyline options: %#v", body)
	}
	if !strings.Contains(fieldMask, "routes.routeLabels") {
		t.Fatalf("route labels missing from field mask: %q", fieldMask)
	}
	if len(preview.Routes) != 2 {
		t.Fatalf("expected two routes, got %#v", preview.Routes)
	}
	if preview.Routes[0].ID != "route-0" ||
		!preview.Routes[0].Recommended ||
		preview.Routes[0].DistanceMeters != 12345 ||
		preview.Routes[0].DurationSeconds != 902 {
		t.Fatalf("unexpected recommended route: %#v", preview.Routes[0])
	}
	if preview.Routes[1].ID != "route-1" ||
		preview.Routes[1].Recommended ||
		preview.Routes[1].DistanceMeters != 11900 ||
		preview.Routes[1].DurationSeconds != 960 {
		t.Fatalf("unexpected alternate route: %#v", preview.Routes[1])
	}
}

func TestPreviewFallsBackToFirstValidRouteAsRecommended(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":1000,"duration":"120s","polyline":{"encodedPolyline":"abc"}},{"distanceMeters":1200,"duration":"140s","routeLabels":["DEFAULT_ROUTE_ALTERNATE"],"polyline":{"encodedPolyline":"def"}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	preview, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Routes) != 2 || !preview.Routes[0].Recommended || preview.Routes[1].Recommended {
		t.Fatalf("expected first route recommendation fallback, got %#v", preview.Routes)
	}
}

func TestPreviewSkipsMalformedAlternativeButKeepsValidDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":1000,"duration":"120s","routeLabels":["DEFAULT_ROUTE"],"polyline":{"encodedPolyline":"abc"}},{"distanceMeters":1200,"duration":"bad","routeLabels":["DEFAULT_ROUTE_ALTERNATE"],"polyline":{"encodedPolyline":"def"}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	preview, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Routes) != 1 || !preview.Routes[0].Recommended {
		t.Fatalf("unexpected routes after malformed alternate: %#v", preview.Routes)
	}
}
func TestPreviewRetriesOneTransientProviderFailure(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":1000,"duration":"120s","routeLabels":["DEFAULT_ROUTE"],"polyline":{"encodedPolyline":"abc"}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	_, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("expected two attempts, got %d", got)
	}
}

func TestPreviewReturnsNotFoundForEmptyRoutes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	_, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != routing.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
