package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

func TestAutocompleteUsesSessionTokenBiasAndNarrowFieldMask(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/places:autocomplete" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Goog-Api-Key") != "server-key" {
			t.Fatal("server API key header missing")
		}
		if got := r.Header.Get("X-Goog-FieldMask"); strings.Contains(got, "*") || !strings.Contains(got, "placeId") || !strings.Contains(got, "text.text") {
			t.Fatalf("unexpected field mask: %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"suggestions":[{"placePrediction":{"placeId":"p1","text":{"text":"Clifton, Karachi"}}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	bias := locationsearch.Point{Latitude: 24.86, Longitude: 67.01}
	items, err := provider.Autocomplete(context.Background(), locationsearch.AutocompleteInput{
		Query: "Clifton", SessionToken: "session-1", Bias: &bias,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].PlaceID != "p1" {
		t.Fatalf("unexpected suggestions: %#v", items)
	}
	if body["sessionToken"] != "session-1" {
		t.Fatalf("session token missing: %#v", body)
	}
	if body["locationBias"] == nil {
		t.Fatalf("location bias missing: %#v", body)
	}
}

func TestDetailsConcludesSessionWithoutProFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/places/p1" {
			t.Fatalf("unexpected details path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("sessionToken") != "session-1" {
			t.Fatalf("missing details session token: %s", r.URL.RawQuery)
		}
		if got := r.Header.Get("X-Goog-FieldMask"); got != "id,formattedAddress,location" {
			t.Fatalf("unexpected details mask: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p1","formattedAddress":"Clifton, Karachi","location":{"latitude":24.82,"longitude":67.03}}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	place, err := provider.Details(context.Background(), "p1", "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if place.Label != "Clifton, Karachi" || place.Location.Latitude != 24.82 {
		t.Fatalf("unexpected place: %#v", place)
	}
}

func TestReverseGeocodePrefersNearestNamedPlaceAndCanonicalLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/places:searchNearby" || r.Method != http.MethodPost {
			t.Fatalf("unexpected nearby request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("X-Goog-FieldMask"); got != "places.id,places.displayName,places.formattedAddress,places.location" {
			t.Fatalf("unexpected nearby mask: %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["rankPreference"] != "DISTANCE" || body["maxResultCount"] != float64(1) {
			t.Fatalf("unexpected nearby body: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"places":[{"id":"named-1","displayName":{"text":"Faisal Mosque"},"formattedAddress":"Shah Faisal Ave, Islamabad","location":{"latitude":33.7295,"longitude":73.0372}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	place, err := provider.ReverseGeocode(context.Background(), locationsearch.Point{Latitude: 33.7294, Longitude: 73.0371})
	if err != nil {
		t.Fatal(err)
	}
	if !place.SnapToPlace || place.PlaceID != "named-1" || place.Location.Latitude != 33.7295 {
		t.Fatalf("expected named-place snap, got %#v", place)
	}
	if !strings.Contains(place.Label, "Faisal Mosque") {
		t.Fatalf("expected named label, got %q", place.Label)
	}
}

func TestReverseGeocodeFallsBackToFormattedAddressAndKeepsPin(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/places:searchNearby":
			_, _ = w.Write([]byte(`{"places":[]}`))
		case "/v4/geocode/location":
			if r.URL.Query().Get("location.latitude") != "24.86" || r.URL.Query().Get("location.longitude") != "67.01" {
				t.Fatalf("unexpected coordinates: %s", r.URL.RawQuery)
			}
			if got := r.Header.Get("X-Goog-FieldMask"); !strings.Contains(got, "formattedAddress") {
				t.Fatalf("missing reverse field mask: %q", got)
			}
			_, _ = w.Write([]byte(`{"results":[{"placeId":"pin-1","formattedAddress":"Unnamed Road, Islamabad","location":{"latitude":24.8602,"longitude":67.0102}}]}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	pin := locationsearch.Point{Latitude: 24.86, Longitude: 67.01}
	place, err := provider.ReverseGeocode(context.Background(), pin)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || place.SnapToPlace || place.Location != pin {
		t.Fatalf("expected address fallback at exact pin, got %#v after %d requests", place, requests)
	}
	if place.Label != "Unnamed Road, Islamabad" {
		t.Fatalf("unexpected fallback label: %q", place.Label)
	}
}
