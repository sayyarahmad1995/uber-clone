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

const (
	recommendedDurationPercent = int64(115)
	recommendedDurationSlack   = 5 * 60
)

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
	if err := validateCandidates(preview.Routes); err != nil {
		return Preview{}, ErrProvider
	}
	recommendRoute(preview.Routes)
	if !preview.Valid() {
		return Preview{}, ErrProvider
	}
	return preview, nil
}

func validateCandidates(routes []Route) error {
	if len(routes) == 0 {
		return ErrProvider
	}
	ids := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		if !route.Valid() {
			return ErrProvider
		}
		if _, exists := ids[route.ID]; exists {
			return ErrProvider
		}
		ids[route.ID] = struct{}{}
	}
	return nil
}

func recommendRoute(routes []Route) {
	fastest := routes[0].DurationSeconds
	for i := range routes {
		routes[i].Recommended = false
		if routes[i].DurationSeconds < fastest {
			fastest = routes[i].DurationSeconds
		}
	}

	percentLimit := (fastest*recommendedDurationPercent + 99) / 100
	slackLimit := fastest + recommendedDurationSlack
	maxDuration := percentLimit
	if slackLimit < maxDuration {
		maxDuration = slackLimit
	}

	selected := -1
	for i := range routes {
		if routes[i].DurationSeconds > maxDuration {
			continue
		}
		if selected == -1 ||
			routes[i].DistanceMeters < routes[selected].DistanceMeters ||
			(routes[i].DistanceMeters == routes[selected].DistanceMeters &&
				routes[i].DurationSeconds < routes[selected].DurationSeconds) {
			selected = i
		}
	}
	if selected == -1 {
		for i := range routes {
			if selected == -1 || routes[i].DurationSeconds < routes[selected].DurationSeconds {
				selected = i
			}
		}
	}
	routes[selected].Recommended = true
}
