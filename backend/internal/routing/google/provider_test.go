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

func TestPreviewRequestsGoogleRecommendedOptimalTrafficRoute(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":12345,"duration":"901.4s","polyline":{"encodedPolyline":"_p~iF~ps|U_ulLnnqC_mqNvxq"}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	route, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if body["travelMode"] != "DRIVE" ||
		body["routingPreference"] != "TRAFFIC_AWARE_OPTIMAL" ||
		body["trafficModel"] != "BEST_GUESS" ||
		body["computeAlternativeRoutes"] != false {
		t.Fatalf("unexpected route options: %#v", body)
	}
	if _, exists := body["requestedReferenceRoutes"]; exists {
		t.Fatalf("shorter-distance route must not be requested: %#v", body)
	}
	if body["polylineQuality"] != "OVERVIEW" || body["polylineEncoding"] != "ENCODED_POLYLINE" {
		t.Fatalf("unexpected polyline options: %#v", body)
	}
	if strings.Contains(fieldMask, "routeLabels") || strings.Contains(fieldMask, "routeToken") {
		t.Fatalf("alternative/reference route fields must not be requested: %q", fieldMask)
	}
	if route.DistanceMeters != 12345 ||
		route.DurationSeconds != 902 ||
		route.EncodedPolyline == "" {
		t.Fatalf("unexpected route: %#v", route)
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
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":1000,"duration":"120s","polyline":{"encodedPolyline":"abc"}}]}`))
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

func TestPreviewRejectsMalformedRecommendedRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":1000,"duration":"bad","polyline":{"encodedPolyline":"abc"}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL)
	_, err := provider.Preview(
		context.Background(),
		routing.Point{Latitude: 24.86, Longitude: 67.01},
		routing.Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err == nil {
		t.Fatal("expected malformed route error")
	}
}
