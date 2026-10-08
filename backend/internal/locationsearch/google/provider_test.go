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

func TestAutocompleteUsesSessionTokenLocalRestrictionAndNarrowFieldMask(t *testing.T) {
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
		RestrictionRadiusMeters: 25000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].PlaceID != "p1" {
		t.Fatalf("unexpected suggestions: %#v", items)
	}
	if body["sessionToken"] != "session-1" {
		t.Fatalf("unexpected autocomplete session: %#v", body)
	}
	if _, exists := body["locationBias"]; exists {
		t.Fatalf("soft location bias must not be used: %#v", body)
	}
	restriction, ok := body["locationRestriction"].(map[string]any)
	if !ok {
		t.Fatalf("local restriction missing: %#v", body)
	}
	circle, ok := restriction["circle"].(map[string]any)
	if !ok || circle["radius"] != float64(25000) {
		t.Fatalf("unexpected local restriction: %#v", restriction)
	}
}

func TestAutocompleteRanksMatchedPredictionsByPopularity(t *testing.T) {
	var detailRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/places:autocomplete":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"suggestions":[
				{"placePrediction":{"placeId":"p1","text":{"text":"Faisal Movers, Islamabad"}}},
				{"placePrediction":{"placeId":"p2","text":{"text":"Faisal Mosque, Islamabad"}}},
				{"placePrediction":{"placeId":"p3","text":{"text":"Faisal Market, Islamabad"}}}
			]}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/places/"):
			detailRequests++
			if got := r.Header.Get("X-Goog-FieldMask"); got != "userRatingCount" {
				t.Fatalf("unexpected popularity field mask: %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/v1/places/p1":
				_, _ = w.Write([]byte(`{"userRatingCount":500}`))
			case "/v1/places/p2":
				_, _ = w.Write([]byte(`{"userRatingCount":12000}`))
			case "/v1/places/p3":
				_, _ = w.Write([]byte(`{"userRatingCount":900}`))
			default:
				t.Fatalf("unexpected popularity path: %s", r.URL.Path)
			}
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	bias := locationsearch.Point{Latitude: 33.6844, Longitude: 73.0479}
	items, err := provider.Autocomplete(context.Background(), locationsearch.AutocompleteInput{
		Query:                   "faisal",
		SessionToken:            "session-1",
		Bias:                    &bias,
		RestrictionRadiusMeters: 25000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 ||
		items[0].PlaceID != "p2" ||
		items[1].PlaceID != "p3" ||
		items[2].PlaceID != "p1" {
		t.Fatalf("unexpected popularity order: %#v", items)
	}

	_, err = provider.Autocomplete(context.Background(), locationsearch.AutocompleteInput{
		Query:                   "faisal",
		SessionToken:            "session-2",
		Bias:                    &bias,
		RestrictionRadiusMeters: 25000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if detailRequests != 3 {
		t.Fatalf("expected cached popularity lookups after first search, got %d requests", detailRequests)
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
		if got := r.Header.Get("X-Goog-FieldMask"); got != "id,displayName.text,formattedAddress,location" {
			t.Fatalf("unexpected details mask: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"p1","displayName":{"text":"Clifton"},"formattedAddress":"Clifton, Karachi","location":{"latitude":24.82,"longitude":67.03}}`))
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

func TestDetailsWithoutAutocompleteSessionOmitsSessionToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/places/poi-1" {
			t.Fatalf("unexpected details path: %s", r.URL.Path)
		}
		if r.URL.Query().Has("sessionToken") {
			t.Fatalf("direct POI details must not fabricate an autocomplete session: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"poi-1","displayName":{"text":"Faisal Mosque"},"formattedAddress":"Shah Faisal Ave, Islamabad","location":{"latitude":33.7295,"longitude":73.0372}}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	place, err := provider.Details(context.Background(), "poi-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if place.PlaceID != "poi-1" || place.Label != "Faisal Mosque" || place.Location.Latitude != 33.7295 {
		t.Fatalf("unexpected POI place: %#v", place)
	}
}

func TestReverseGeocodeReturnsReadableAddressAtExactPin(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/v4/geocode/location" {
			t.Fatalf("reverse geocode must not query Nearby Search: %s", r.URL.Path)
		}
		if r.URL.Query().Get("location.latitude") != "24.86" || r.URL.Query().Get("location.longitude") != "67.01" {
			t.Fatalf("unexpected coordinates: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"placeId":"pin-1","formattedAddress":"Unnamed Road, Islamabad","location":{"latitude":24.8602,"longitude":67.0102}}]}`))
	}))
	defer server.Close()

	provider := newProvider("server-key", server.Client(), server.URL, server.URL)
	pin := locationsearch.Point{Latitude: 24.86, Longitude: 67.01}
	place, err := provider.ReverseGeocode(context.Background(), pin)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 || place.Location != pin {
		t.Fatalf("expected one geocode request and exact pin retention, got %#v after %d requests", place, requests)
	}
	if place.Label != "Unnamed Road, Islamabad" {
		t.Fatalf("unexpected label: %q", place.Label)
	}
}
