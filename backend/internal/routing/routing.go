package routing

import (
	"context"
	"errors"
	"math"
	"strings"
)

var (
	ErrUnavailable  = errors.New("routing is unavailable")
	ErrInvalidInput = errors.New("routing input is invalid")
	ErrNotFound     = errors.New("route was not found")
	ErrProvider     = errors.New("routing provider failed")
)

type Point struct {
	Latitude  float64
	Longitude float64
}

func (p Point) Valid() bool {
	return !math.IsNaN(p.Latitude) &&
		!math.IsNaN(p.Longitude) &&
		!math.IsInf(p.Latitude, 0) &&
		!math.IsInf(p.Longitude, 0) &&
		p.Latitude >= -90 &&
		p.Latitude <= 90 &&
		p.Longitude >= -180 &&
		p.Longitude <= 180
}

type Route struct {
	ID              string
	Recommended     bool
	DistanceMeters  int64
	DurationSeconds int64
	EncodedPolyline string
}

func (r Route) Valid() bool {
	return strings.TrimSpace(r.ID) != "" &&
		r.DistanceMeters > 0 &&
		r.DurationSeconds > 0 &&
		strings.TrimSpace(r.EncodedPolyline) != ""
}

type Preview struct {
	Routes []Route
}

func (p Preview) Valid() bool {
	if len(p.Routes) == 0 {
		return false
	}
	ids := make(map[string]struct{}, len(p.Routes))
	recommended := 0
	for _, route := range p.Routes {
		if !route.Valid() {
			return false
		}
		id := strings.TrimSpace(route.ID)
		if _, exists := ids[id]; exists {
			return false
		}
		ids[id] = struct{}{}
		if route.Recommended {
			recommended++
		}
	}
	return recommended == 1
}

type Provider interface {
	Preview(context.Context, Point, Point) (Preview, error)
}

type Previewer interface {
	Preview(context.Context, Point, Point) (Preview, error)
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) Service {
	return Service{provider: provider}
}

func (s Service) Preview(ctx context.Context, pickup, destination Point) (Preview, error) {
	if s.provider == nil {
		return Preview{}, ErrUnavailable
	}
	if !pickup.Valid() || !destination.Valid() || pickup == destination {
		return Preview{}, ErrInvalidInput
	}
	preview, err := s.provider.Preview(ctx, pickup, destination)
	if err != nil {
		return Preview{}, err
	}
	for i := range preview.Routes {
		preview.Routes[i].ID = strings.TrimSpace(preview.Routes[i].ID)
		preview.Routes[i].EncodedPolyline = strings.TrimSpace(preview.Routes[i].EncodedPolyline)
	}
	if !preview.Valid() {
		return Preview{}, ErrProvider
	}
	return preview, nil
}
