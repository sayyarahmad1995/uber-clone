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
	DistanceMeters  int64
	DurationSeconds int64
	EncodedPolyline string
}

func (r Route) Valid() bool {
	return r.DistanceMeters > 0 &&
		r.DurationSeconds > 0 &&
		strings.TrimSpace(r.EncodedPolyline) != ""
}

type Provider interface {
	Preview(context.Context, Point, Point) (Route, error)
}

type Previewer interface {
	Preview(context.Context, Point, Point) (Route, error)
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) Service {
	return Service{provider: provider}
}

func (s Service) Preview(ctx context.Context, pickup, destination Point) (Route, error) {
	if s.provider == nil {
		return Route{}, ErrUnavailable
	}
	if !pickup.Valid() || !destination.Valid() || pickup == destination {
		return Route{}, ErrInvalidInput
	}
	route, err := s.provider.Preview(ctx, pickup, destination)
	if err != nil {
		return Route{}, err
	}
	route.EncodedPolyline = strings.TrimSpace(route.EncodedPolyline)
	if !route.Valid() {
		return Route{}, ErrProvider
	}
	return route, nil
}
