package routing

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	preview Preview
	err     error
}

func (f fakeProvider) Preview(context.Context, Point, Point) (Preview, error) {
	return f.preview, f.err
}

func validPreview() Preview {
	return Preview{Routes: []Route{{
		ID:              "route-0",
		Recommended:     true,
		DistanceMeters:  100,
		DurationSeconds: 20,
		EncodedPolyline: "abc",
	}}}
}

func TestServiceValidatesCoordinatesAndDistinctEndpoints(t *testing.T) {
	service := NewService(fakeProvider{preview: validPreview()})

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

func TestServiceRejectsMalformedProviderPreview(t *testing.T) {
	cases := []struct {
		name    string
		preview Preview
	}{
		{"empty routes", Preview{}},
		{"malformed route", Preview{Routes: []Route{{ID: "route-0", Recommended: true, DistanceMeters: 100}}}},
		{"no recommended route", Preview{Routes: []Route{{ID: "route-0", DistanceMeters: 100, DurationSeconds: 20, EncodedPolyline: "abc"}}}},
		{"multiple recommended routes", Preview{Routes: []Route{
			{ID: "route-0", Recommended: true, DistanceMeters: 100, DurationSeconds: 20, EncodedPolyline: "abc"},
			{ID: "route-1", Recommended: true, DistanceMeters: 120, DurationSeconds: 22, EncodedPolyline: "def"},
		}}},
		{"duplicate route IDs", Preview{Routes: []Route{
			{ID: "route-0", Recommended: true, DistanceMeters: 100, DurationSeconds: 20, EncodedPolyline: "abc"},
			{ID: "route-0", DistanceMeters: 120, DurationSeconds: 22, EncodedPolyline: "def"},
		}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(fakeProvider{preview: tc.preview})
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
