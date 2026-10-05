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

func previewRoute(
	id string,
	distanceMeters int64,
	durationSeconds int64,
	recommended bool,
) Route {
	return Route{
		ID:              id,
		Recommended:     recommended,
		DistanceMeters:  distanceMeters,
		DurationSeconds: durationSeconds,
		EncodedPolyline: "poly-" + id,
	}
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
		{"malformed route", Preview{Routes: []Route{{
			ID:              "route-0",
			DistanceMeters:  100,
			DurationSeconds: 20,
		}}}},
		{"duplicate route IDs", Preview{Routes: []Route{
			previewRoute("route-0", 100, 20, true),
			previewRoute("route-0", 120, 22, false),
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

func TestServiceRecommendsShortestRouteWithinTrafficTolerance(t *testing.T) {
	service := NewService(fakeProvider{preview: Preview{Routes: []Route{
		previewRoute("fastest", 8000, 18*60, true),
		previewRoute("balanced", 6900, 20*60, false),
		previewRoute("too-slow", 5800, 22*60, false),
	}}})

	preview, err := service.Preview(
		context.Background(),
		Point{Latitude: 24.86, Longitude: 67.01},
		Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Routes[1].Recommended {
		t.Fatalf("expected balanced route to be recommended: %#v", preview.Routes)
	}
	if preview.Routes[0].Recommended || preview.Routes[2].Recommended {
		t.Fatalf("expected exactly one HiGO recommendation: %#v", preview.Routes)
	}
}

func TestServiceAppliesFiveMinuteAbsoluteTrafficCap(t *testing.T) {
	service := NewService(fakeProvider{preview: Preview{Routes: []Route{
		previewRoute("fastest", 30000, 60*60, true),
		previewRoute("shorter-but-six-minutes-slower", 25000, 66*60, false),
		previewRoute("eligible", 27000, 64*60, false),
	}}})

	preview, err := service.Preview(
		context.Background(),
		Point{Latitude: 24.86, Longitude: 67.01},
		Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Routes[2].Recommended {
		t.Fatalf("expected four-minute-slower shorter route: %#v", preview.Routes)
	}
}

func TestServiceFallsBackToFastestWhenShorterRoutesExceedTolerance(t *testing.T) {
	service := NewService(fakeProvider{preview: Preview{Routes: []Route{
		previewRoute("fastest", 8000, 10*60, false),
		previewRoute("shorter", 6000, 13*60, true),
	}}})

	preview, err := service.Preview(
		context.Background(),
		Point{Latitude: 24.86, Longitude: 67.01},
		Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Routes[0].Recommended || preview.Routes[1].Recommended {
		t.Fatalf("expected fastest route fallback: %#v", preview.Routes)
	}
}

func TestServiceOverridesProviderRecommendationUsingHiGOPolicy(t *testing.T) {
	service := NewService(fakeProvider{preview: Preview{Routes: []Route{
		previewRoute("google-default", 9000, 20*60, true),
		previewRoute("shorter", 8000, 21*60, false),
	}}})

	preview, err := service.Preview(
		context.Background(),
		Point{Latitude: 24.86, Longitude: 67.01},
		Point{Latitude: 24.90, Longitude: 67.05},
	)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Routes[0].Recommended || !preview.Routes[1].Recommended {
		t.Fatalf("expected HiGO policy to select shorter qualifying route: %#v", preview.Routes)
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
