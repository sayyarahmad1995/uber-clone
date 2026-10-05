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

func (f fakeProvider) Preview(context.Context, Point, Point) (Route, error) {
	return f.route, f.err
}

func validRoute() Route {
	return Route{
		DistanceMeters:  100,
		DurationSeconds: 20,
		EncodedPolyline: "abc",
	}
}

func TestServiceValidatesCoordinatesAndDistinctEndpoints(t *testing.T) {
	service := NewService(fakeProvider{route: validRoute()})

	cases := []struct {
		name        string
		pickup      Point
		destination Point
	}{
		{"invalid pickup", Point{Latitude: 91}, Point{Latitude: 24, Longitude: 67}},
		{"invalid destination", Point{Latitude: 24, Longitude: 67}, Point{Longitude: 181}},
		{"same endpoint", Point{Latitude: 24, Longitude: 67}, Point{Latitude: 24, Longitude: 67}},
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
				Point{Latitude: 24.86, Longitude: 67.01},
				Point{Latitude: 24.90, Longitude: 67.05},
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
		Point{Latitude: 24.86, Longitude: 67.01},
		Point{Latitude: 24.90, Longitude: 67.05},
	)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
