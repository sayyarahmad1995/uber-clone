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
	fieldMask         = "routes.distanceMeters,routes.duration,routes.polyline.encodedPolyline"
	maxAttempts       = 2
	retryDelay        = 100 * time.Millisecond
)

type Provider struct {
	apiKey string
	client *http.Client
	base   string
}

func New(apiKey string) *Provider {
	return newProvider(
		strings.TrimSpace(apiKey),
		&http.Client{Timeout: 5 * time.Second},
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

func (p *Provider) Preview(ctx context.Context, pickup, destination routing.Point) (routing.Route, error) {
	if p.apiKey == "" {
		return routing.Route{}, routing.ErrUnavailable
	}

	body := map[string]any{
		"origin": waypoint(pickup),
		"destination": waypoint(destination),
		"travelMode": "DRIVE",
		"computeAlternativeRoutes": false,
		"polylineQuality": "OVERVIEW",
		"polylineEncoding": "ENCODED_POLYLINE",
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return routing.Route{}, fmt.Errorf("%w: encode route request", routing.ErrProvider)
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
			return routing.Route{}, fmt.Errorf("%w: build route request", routing.ErrProvider)
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
				return routing.Route{}, ctx.Err()
			}
			return routing.Route{}, err
		}
		if attempt+1 < maxAttempts {
			timer := time.NewTimer(retryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return routing.Route{}, ctx.Err()
			case <-timer.C:
			}
		}
	}
	if err != nil {
		return routing.Route{}, fmt.Errorf("%w: compute route request", routing.ErrProvider)
	}
	if resp == nil {
		return routing.Route{}, fmt.Errorf("%w: empty route response", routing.ErrProvider)
	}
	defer resp.Body.Close()
	if err := providerStatus(resp.StatusCode); err != nil {
		return routing.Route{}, err
	}

	var payload struct {
		Routes []struct {
			DistanceMeters int64  `json:"distanceMeters"`
			Duration       string `json:"duration"`
			Polyline       struct {
				EncodedPolyline string `json:"encodedPolyline"`
			} `json:"polyline"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return routing.Route{}, fmt.Errorf("%w: decode route response", routing.ErrProvider)
	}
	if len(payload.Routes) == 0 {
		return routing.Route{}, routing.ErrNotFound
	}

	first := payload.Routes[0]
	duration, err := time.ParseDuration(strings.TrimSpace(first.Duration))
	if err != nil || duration <= 0 {
		return routing.Route{}, fmt.Errorf("%w: invalid route duration", routing.ErrProvider)
	}
	route := routing.Route{
		DistanceMeters:  first.DistanceMeters,
		DurationSeconds: int64(math.Ceil(duration.Seconds())),
		EncodedPolyline: strings.TrimSpace(first.Polyline.EncodedPolyline),
	}
	if !route.Valid() {
		return routing.Route{}, fmt.Errorf("%w: incomplete route", routing.ErrProvider)
	}
	return route, nil
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
