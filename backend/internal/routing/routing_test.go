package routing

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	route Route
	err   error
}

func (f fakeProvider) Preview(context.Context, Endpoint, Endpoint) (Route, error) {
	return f.route, f.err
}

func validRoute() Route {
	return Route{
		DistanceMeters:  100,
		DurationSeconds: 20,
		EncodedPolyline: "abc",
	}
}

func endpoint(latitude, longitude float64, placeID string) Endpoint {
	return Endpoint{
		Point:   Point{Latitude: latitude, Longitude: longitude},
		PlaceID: placeID,
	}
}

func TestServiceValidatesCoordinatesAndDistinctEndpoints(t *testing.T) {
	service := NewService(fakeProvider{route: validRoute()})

	cases := []struct {
		name        string
		pickup      Endpoint
		destination Endpoint
	}{
		{"invalid pickup", endpoint(91, 0, ""), endpoint(24, 67, "")},
		{"invalid destination", endpoint(24, 67, ""), endpoint(0, 181, "")},
		{"same endpoint", endpoint(24, 67, ""), endpoint(24, 67, "place-b")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.Preview(context.Background(), tc.pickup, tc.destination)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
		})
	}
}

func TestServiceAcceptsOptionalPlaceIDs(t *testing.T) {
	service := NewService(fakeProvider{route: validRoute()})
	_, err := service.Preview(
		context.Background(),
		endpoint(24.86, 67.01, " pickup-place "),
		endpoint(24.90, 67.05, "destination-place"),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceRejectsMalformedProviderRoute(t *testing.T) {
	cases := []struct {
		name  string
		route Route
	}{
		{"zero distance", Route{DurationSeconds: 20, EncodedPolyline: "abc"}},
		{"zero duration", Route{DistanceMeters: 100, EncodedPolyline: "abc"}},
		{"missing polyline", Route{DistanceMeters: 100, DurationSeconds: 20}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(fakeProvider{route: tc.route})
			_, err := service.Preview(
				context.Background(),
				endpoint(24.86, 67.01, ""),
				endpoint(24.90, 67.05, ""),
			)
			if !errors.Is(err, ErrProvider) {
				t.Fatalf("expected provider error, got %v", err)
			}
		})
	}
}

func TestServicePropagatesProviderError(t *testing.T) {
	service := NewService(fakeProvider{err: ErrNotFound})
	_, err := service.Preview(
		context.Background(),
		endpoint(24.86, 67.01, ""),
		endpoint(24.90, 67.05, ""),
	)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
