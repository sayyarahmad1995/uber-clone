package locationsearch

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSearchPolicyService struct {
	policy SearchPolicy
	err    error
}

func (f fakeSearchPolicyService) Load(context.Context) (SearchPolicy, error) {
	return f.policy, f.err
}

func (f fakeSearchPolicyService) Update(
	context.Context,
	SearchPolicy,
	string,
) (SearchPolicy, error) {
	return SearchPolicy{}, nil
}

type fakeProvider struct {
	autocompleteInput AutocompleteInput
	reversePoint      Point
}

func (f *fakeProvider) Autocomplete(_ context.Context, input AutocompleteInput) ([]Suggestion, error) {
	f.autocompleteInput = input
	return []Suggestion{{PlaceID: "place-1", Label: "Test place"}}, nil
}

func (f *fakeProvider) Details(context.Context, string, string) (Place, error) {
	return Place{PlaceID: "place-1", Label: "Test place", Location: Point{Latitude: 24, Longitude: 67}}, nil
}

func (f *fakeProvider) ReverseGeocode(_ context.Context, point Point) (Place, error) {
	f.reversePoint = point
	return Place{Label: "Pinned place", Location: point}, nil
}

func TestServiceValidatesAutocompleteSessionAndBias(t *testing.T) {
	provider := &fakeProvider{}
	service := NewService(provider)
	bias := Point{Latitude: 24.86, Longitude: 67.01}

	items, err := service.Autocomplete(context.Background(), AutocompleteInput{
		Query:        "  Clifton  ",
		SessionToken: " session ",
		Bias:         &bias,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 ||
		provider.autocompleteInput.Query != "Clifton" ||
		provider.autocompleteInput.SessionToken != "session" ||
		provider.autocompleteInput.RestrictionRadiusMeters != DefaultAutocompleteRadiusMeters {
		t.Fatalf("unexpected autocomplete input: %#v", provider.autocompleteInput)
	}
}

func TestServiceUsesConfiguredAutocompleteRadius(t *testing.T) {
	provider := &fakeProvider{}
	service := NewServiceWithPolicy(provider, fakeSearchPolicyService{
		policy: SearchPolicy{
			AutocompleteRadiusMeters: 25000,
			UpdatedAt:                time.Now().UTC(),
			UpdatedBy:                "admin",
		},
	})
	bias := Point{Latitude: 33.6844, Longitude: 73.0479}

	_, err := service.Autocomplete(context.Background(), AutocompleteInput{
		Query:        "faisal",
		SessionToken: "session",
		Bias:         &bias,
	})
	if err != nil {
		t.Fatal(err)
	}
	if provider.autocompleteInput.RestrictionRadiusMeters != 25000 {
		t.Fatalf("unexpected configured radius: %#v", provider.autocompleteInput)
	}
}

func TestServiceRejectsInvalidRequestsBeforeProvider(t *testing.T) {
	service := NewService(&fakeProvider{})
	if _, err := service.Autocomplete(context.Background(), AutocompleteInput{Query: "x", SessionToken: "token"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid autocomplete, got %v", err)
	}
	if _, err := service.Details(context.Background(), "", "token"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid details, got %v", err)
	}
	if _, err := service.ReverseGeocode(context.Background(), Point{Latitude: 91}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid reverse geocode, got %v", err)
	}
}

func TestServiceAllowsDirectPlaceDetailsWithoutAutocompleteSession(t *testing.T) {
	service := NewService(&fakeProvider{})

	place, err := service.Details(context.Background(), "place-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if place.PlaceID != "place-1" {
		t.Fatalf("unexpected place: %#v", place)
	}
}

func TestServiceReverseGeocodePassesExactPointToProvider(t *testing.T) {
	provider := &fakeProvider{}
	service := NewService(provider)
	point := Point{Latitude: 33.7294, Longitude: 73.0371}

	place, err := service.ReverseGeocode(context.Background(), point)
	if err != nil {
		t.Fatal(err)
	}
	if provider.reversePoint != point || place.Location != point {
		t.Fatalf("reverse geocode must preserve exact point: provider=%#v place=%#v", provider.reversePoint, place)
	}
}
