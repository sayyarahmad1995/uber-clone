package locationsearch

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	autocompleteInput AutocompleteInput
	reverseRadius     int64
}

func (f *fakeProvider) Autocomplete(_ context.Context, input AutocompleteInput) ([]Suggestion, error) {
	f.autocompleteInput = input
	return []Suggestion{{PlaceID: "place-1", Label: "Test place"}}, nil
}

func (f *fakeProvider) Details(context.Context, string, string) (Place, error) {
	return Place{PlaceID: "place-1", Label: "Test place", Location: Point{Latitude: 24, Longitude: 67}}, nil
}

func (f *fakeProvider) ReverseGeocode(_ context.Context, point Point, radius int64) (Place, error) {
	f.reverseRadius = radius
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
	if len(items) != 1 || provider.autocompleteInput.Query != "Clifton" || provider.autocompleteInput.SessionToken != "session" {
		t.Fatalf("unexpected autocomplete input: %#v", provider.autocompleteInput)
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

type fakePolicyService struct {
	policy Policy
	err    error
}

func (f *fakePolicyService) Load(context.Context) (Policy, error) {
	return f.policy, f.err
}

func (f *fakePolicyService) Update(_ context.Context, policy Policy, _ string) (Policy, error) {
	f.policy = policy
	return policy, nil
}

func TestServiceLoadsCurrentSnapRadiusForEachReverseGeocode(t *testing.T) {
	provider := &fakeProvider{}
	policy := &fakePolicyService{policy: Policy{NamedPlaceSnapRadiusMeters: 5}}
	service := NewService(provider, policy)
	point := Point{Latitude: 33.7, Longitude: 73.0}

	if _, err := service.ReverseGeocode(context.Background(), point); err != nil {
		t.Fatal(err)
	}
	if provider.reverseRadius != 5 {
		t.Fatalf("expected initial 5 meter radius, got %d", provider.reverseRadius)
	}

	policy.policy.NamedPlaceSnapRadiusMeters = 9
	if _, err := service.ReverseGeocode(context.Background(), point); err != nil {
		t.Fatal(err)
	}
	if provider.reverseRadius != 9 {
		t.Fatalf("expected updated 9 meter radius without restart, got %d", provider.reverseRadius)
	}
}

func TestServiceUsesDefaultSnapRadiusWithoutPolicyService(t *testing.T) {
	provider := &fakeProvider{}
	service := NewService(provider)

	if _, err := service.ReverseGeocode(
		context.Background(),
		Point{Latitude: 33.7, Longitude: 73.0},
	); err != nil {
		t.Fatal(err)
	}
	if provider.reverseRadius != DefaultNamedPlaceSnapRadiusMeters {
		t.Fatalf("expected default radius %d, got %d", DefaultNamedPlaceSnapRadiusMeters, provider.reverseRadius)
	}
}

func TestServiceReturnsUnavailableWhenPolicyCannotLoad(t *testing.T) {
	provider := &fakeProvider{}
	policy := &fakePolicyService{err: errors.New("database unavailable")}
	service := NewService(provider, policy)

	_, err := service.ReverseGeocode(
		context.Background(),
		Point{Latitude: 33.7, Longitude: 73.0},
	)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}
