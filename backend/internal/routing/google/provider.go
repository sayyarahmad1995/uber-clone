package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
)

const (
	defaultRoutesBase = "https://routes.googleapis.com"
	fieldMask         = "routes.distanceMeters,routes.duration,routes.polyline.encodedPolyline,routes.routeLabels,routes.routeToken"
	maxAttempts       = 2
	retryDelay        = 100 * time.Millisecond
	providerTimeout   = 8 * time.Second
)

type Provider struct {
	apiKey string
	client *http.Client
	base   string
}

func New(apiKey string) *Provider {
	return newProvider(
		strings.TrimSpace(apiKey),
		&http.Client{Timeout: providerTimeout},
		defaultRoutesBase,
	)
}

func newProvider(apiKey string, client *http.Client, base string) *Provider {
	return &Provider{
		apiKey: strings.TrimSpace(apiKey),
		client: client,
		base:   strings.TrimRight(base, "/"),
	}
}

func (p *Provider) Preview(ctx context.Context, pickup, destination routing.Point) (routing.Preview, error) {
	if p.apiKey == "" {
		return routing.Preview{}, routing.ErrUnavailable
	}

	body := map[string]any{
		"origin":                   waypoint(pickup),
		"destination":              waypoint(destination),
		"travelMode":               "DRIVE",
		"routingPreference":        "TRAFFIC_AWARE_OPTIMAL",
		"trafficModel":             "BEST_GUESS",
		"computeAlternativeRoutes": true,
		"requestedReferenceRoutes": []string{"SHORTER_DISTANCE"},
		"polylineQuality":          "OVERVIEW",
		"polylineEncoding":         "ENCODED_POLYLINE",
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return routing.Preview{}, fmt.Errorf("%w: encode route request", routing.ErrProvider)
	}

	var resp *http.Response
	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, buildErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			p.base+"/directions/v2:computeRoutes",
			bytes.NewReader(encoded),
		)
		if buildErr != nil {
			return routing.Preview{}, fmt.Errorf("%w: build route request", routing.ErrProvider)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Goog-Api-Key", p.apiKey)
		req.Header.Set("X-Goog-FieldMask", fieldMask)

		resp, err = p.client.Do(req)
		if err == nil && !retryableStatus(resp.StatusCode) {
			break
		}
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			resp = nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			if ctx.Err() != nil {
				return routing.Preview{}, ctx.Err()
			}
			return routing.Preview{}, err
		}
		if attempt+1 < maxAttempts {
			timer := time.NewTimer(retryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return routing.Preview{}, ctx.Err()
			case <-timer.C:
			}
		}
	}
	if err != nil {
		return routing.Preview{}, fmt.Errorf("%w: compute route request", routing.ErrProvider)
	}
	if resp == nil {
		return routing.Preview{}, fmt.Errorf("%w: empty route response", routing.ErrProvider)
	}
	defer resp.Body.Close()
	if err := providerStatus(resp.StatusCode); err != nil {
		return routing.Preview{}, err
	}

	var payload struct {
		Routes []struct {
			DistanceMeters int64    `json:"distanceMeters"`
			Duration       string   `json:"duration"`
			RouteLabels    []string `json:"routeLabels"`
			Polyline       struct {
				EncodedPolyline string `json:"encodedPolyline"`
			} `json:"polyline"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return routing.Preview{}, fmt.Errorf("%w: decode route response", routing.ErrProvider)
	}
	if len(payload.Routes) == 0 {
		return routing.Preview{}, routing.ErrNotFound
	}

	routes := make([]routing.Route, 0, len(payload.Routes))
	seenPolylines := make(map[string]struct{}, len(payload.Routes))
	recommendedIndex := -1
	for _, candidate := range payload.Routes {
		duration, parseErr := time.ParseDuration(strings.TrimSpace(candidate.Duration))
		if parseErr != nil || duration <= 0 {
			continue
		}
		polyline := strings.TrimSpace(candidate.Polyline.EncodedPolyline)
		if _, exists := seenPolylines[polyline]; exists {
			continue
		}
		route := routing.Route{
			ID:              fmt.Sprintf("route-%d", len(routes)),
			Recommended:     hasRouteLabel(candidate.RouteLabels, "DEFAULT_ROUTE"),
			DistanceMeters:  candidate.DistanceMeters,
			DurationSeconds: int64(math.Ceil(duration.Seconds())),
			EncodedPolyline: polyline,
		}
		if !route.Valid() {
			continue
		}
		seenPolylines[polyline] = struct{}{}
		if route.Recommended && recommendedIndex == -1 {
			recommendedIndex = len(routes)
		} else {
			route.Recommended = false
		}
		routes = append(routes, route)
	}
	if len(routes) == 0 {
		return routing.Preview{}, fmt.Errorf("%w: no complete routes", routing.ErrProvider)
	}
	if recommendedIndex == -1 {
		routes[0].Recommended = true
	}
	return routing.Preview{Routes: routes}, nil
}

func hasRouteLabel(labels []string, expected string) bool {
	for _, label := range labels {
		if strings.EqualFold(strings.TrimSpace(label), expected) {
			return true
		}
	}
	return false
}

func waypoint(point routing.Point) map[string]any {
	return map[string]any{
		"location": map[string]any{
			"latLng": map[string]any{
				"latitude":  point.Latitude,
				"longitude": point.Longitude,
			},
		},
	}
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func providerStatus(status int) error {
	switch {
	case status >= 200 && status < 300:
		return nil
	case status == http.StatusBadRequest:
		return routing.ErrInvalidInput
	case status == http.StatusNotFound:
		return routing.ErrNotFound
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return routing.ErrUnavailable
	default:
		return routing.ErrProvider
	}
}
