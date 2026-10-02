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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

const (
	defaultPlacesBase  = "https://places.googleapis.com"
	defaultGeocodeBase = "https://geocode.googleapis.com"
)

type Provider struct {
	apiKey      string
	client      *http.Client
	placesBase  string
	geocodeBase string
}

func New(apiKey string) *Provider {
	return newProvider(
		strings.TrimSpace(apiKey),
		&http.Client{Timeout: 5 * time.Second},
		defaultPlacesBase,
		defaultGeocodeBase,
	)
}

func newProvider(apiKey string, client *http.Client, placesBase, geocodeBase string) *Provider {
	return &Provider{
		apiKey:      strings.TrimSpace(apiKey),
		client:      client,
		placesBase:  strings.TrimRight(placesBase, "/"),
		geocodeBase: strings.TrimRight(geocodeBase, "/"),
	}
}

func (p *Provider) Autocomplete(ctx context.Context, input locationsearch.AutocompleteInput) ([]locationsearch.Suggestion, error) {
	if p.apiKey == "" {
		return nil, locationsearch.ErrUnavailable
	}
	body := map[string]any{
		"input":        input.Query,
		"sessionToken": input.SessionToken,
	}
	if input.Bias != nil {
		body["locationBias"] = map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  input.Bias.Latitude,
					"longitude": input.Bias.Longitude,
				},
				"radius": 50000.0,
			},
		}
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%w: encode autocomplete request", locationsearch.ErrProvider)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.placesBase+"/v1/places:autocomplete",
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: build autocomplete request", locationsearch.ErrProvider)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set(
		"X-Goog-FieldMask",
		"suggestions.placePrediction.placeId,suggestions.placePrediction.text.text",
	)

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: autocomplete request", locationsearch.ErrProvider)
	}
	defer resp.Body.Close()
	if err := providerStatus(resp.StatusCode); err != nil {
		return nil, err
	}

	var payload struct {
		Suggestions []struct {
			PlacePrediction *struct {
				PlaceID string `json:"placeId"`
				Text    struct {
					Text string `json:"text"`
				} `json:"text"`
			} `json:"placePrediction"`
		} `json:"suggestions"`
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return nil, err
	}
	items := make([]locationsearch.Suggestion, 0, len(payload.Suggestions))
	for _, suggestion := range payload.Suggestions {
		prediction := suggestion.PlacePrediction
		if prediction == nil {
			continue
		}
		placeID := strings.TrimSpace(prediction.PlaceID)
		label := strings.TrimSpace(prediction.Text.Text)
		if placeID == "" || label == "" {
			continue
		}
		items = append(items, locationsearch.Suggestion{PlaceID: placeID, Label: label})
	}
	return items, nil
}

func (p *Provider) Details(ctx context.Context, placeID, sessionToken string) (locationsearch.Place, error) {
	if p.apiKey == "" {
		return locationsearch.Place{}, locationsearch.ErrUnavailable
	}
	endpoint, err := url.Parse(p.placesBase + "/v1/places/" + url.PathEscape(placeID))
	if err != nil {
		return locationsearch.Place{}, fmt.Errorf("%w: build place details URL", locationsearch.ErrProvider)
	}
	query := endpoint.Query()
	query.Set("sessionToken", sessionToken)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return locationsearch.Place{}, fmt.Errorf("%w: build place details request", locationsearch.ErrProvider)
	}
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set("X-Goog-FieldMask", "id,formattedAddress,location")

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return locationsearch.Place{}, err
		}
		return locationsearch.Place{}, fmt.Errorf("%w: place details request", locationsearch.ErrProvider)
	}
	defer resp.Body.Close()
	if err := providerStatus(resp.StatusCode); err != nil {
		return locationsearch.Place{}, err
	}

	var payload struct {
		ID               string `json:"id"`
		FormattedAddress string `json:"formattedAddress"`
		Location         struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return locationsearch.Place{}, err
	}
	place := locationsearch.Place{
		PlaceID:     strings.TrimSpace(payload.ID),
		Label:       strings.TrimSpace(payload.FormattedAddress),
		SnapToPlace: true,
		Location: locationsearch.Point{
			Latitude:  payload.Location.Latitude,
			Longitude: payload.Location.Longitude,
		},
	}
	if place.PlaceID == "" || place.Label == "" || !place.Location.Valid() {
		return locationsearch.Place{}, locationsearch.ErrNotFound
	}
	return place, nil
}

func (p *Provider) ReverseGeocode(ctx context.Context, point locationsearch.Point, namedPlaceSnapRadiusMeters int64) (locationsearch.Place, error) {
	if p.apiKey == "" {
		return locationsearch.Place{}, locationsearch.ErrUnavailable
	}
	if namedPlaceSnapRadiusMeters < locationsearch.MinNamedPlaceSnapRadiusMeters ||
		namedPlaceSnapRadiusMeters > locationsearch.MaxNamedPlaceSnapRadiusMeters {
		return locationsearch.Place{}, locationsearch.ErrInvalidInput
	}
	if place, found, err := p.nearbyNamedPlace(ctx, point, namedPlaceSnapRadiusMeters); err != nil {
		return locationsearch.Place{}, err
	} else if found {
		return place, nil
	}
	return p.reverseGeocodeAddress(ctx, point)
}

func (p *Provider) nearbyNamedPlace(ctx context.Context, point locationsearch.Point, radiusMeters int64) (locationsearch.Place, bool, error) {
	body := map[string]any{
		"maxResultCount": 5,
		"rankPreference": "POPULARITY",
		"locationRestriction": map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  point.Latitude,
					"longitude": point.Longitude,
				},
				"radius": float64(radiusMeters),
			},
		},
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return locationsearch.Place{}, false, fmt.Errorf("%w: encode nearby search request", locationsearch.ErrProvider)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.placesBase+"/v1/places:searchNearby", bytes.NewReader(encoded))
	if err != nil {
		return locationsearch.Place{}, false, fmt.Errorf("%w: build nearby search request", locationsearch.ErrProvider)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set(
		"X-Goog-FieldMask",
		"places.id,places.displayName,places.formattedAddress,places.location,places.primaryType,places.types,places.pureServiceAreaBusiness,places.businessStatus",
	)

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return locationsearch.Place{}, false, err
		}
		return locationsearch.Place{}, false, fmt.Errorf("%w: nearby search request", locationsearch.ErrProvider)
	}
	defer resp.Body.Close()
	if err := providerStatus(resp.StatusCode); err != nil {
		return locationsearch.Place{}, false, err
	}

	var payload struct {
		Places []struct {
			ID          string `json:"id"`
			DisplayName struct {
				Text string `json:"text"`
			} `json:"displayName"`
			FormattedAddress        string   `json:"formattedAddress"`
			PrimaryType             string   `json:"primaryType"`
			Types                   []string `json:"types"`
			PureServiceAreaBusiness bool     `json:"pureServiceAreaBusiness"`
			BusinessStatus          string   `json:"businessStatus"`
			Location                struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
		} `json:"places"`
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return locationsearch.Place{}, false, err
	}
	nearestDistance := math.Inf(1)
	var nearest locationsearch.Place
	for _, candidate := range payload.Places {
		name := strings.TrimSpace(candidate.DisplayName.Text)
		location := locationsearch.Point{Latitude: candidate.Location.Latitude, Longitude: candidate.Location.Longitude}
		if name == "" || !location.Valid() {
			continue
		}
		if !suitableNamedPlaceCandidate(
			candidate.PrimaryType,
			candidate.Types,
			candidate.PureServiceAreaBusiness,
			candidate.BusinessStatus,
		) {
			continue
		}
		distance := distanceMeters(point, location)
		if distance > float64(radiusMeters) || distance >= nearestDistance {
			continue
		}
		address := strings.TrimSpace(candidate.FormattedAddress)
		label := name
		if address != "" && !strings.Contains(strings.ToLower(address), strings.ToLower(name)) {
			label += ", " + address
		}
		nearestDistance = distance
		nearest = locationsearch.Place{
			PlaceID:     strings.TrimSpace(candidate.ID),
			Label:       label,
			Location:    location,
			SnapToPlace: true,
		}
	}
	if nearestDistance < math.Inf(1) {
		return nearest, true, nil
	}
	return locationsearch.Place{}, false, nil
}

func (p *Provider) reverseGeocodeAddress(ctx context.Context, point locationsearch.Point) (locationsearch.Place, error) {
	endpoint, err := url.Parse(p.geocodeBase + "/v4/geocode/location")
	if err != nil {
		return locationsearch.Place{}, fmt.Errorf("%w: build reverse geocode URL", locationsearch.ErrProvider)
	}
	query := endpoint.Query()
	query.Set("location.latitude", strconv.FormatFloat(point.Latitude, 'f', -1, 64))
	query.Set("location.longitude", strconv.FormatFloat(point.Longitude, 'f', -1, 64))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return locationsearch.Place{}, fmt.Errorf("%w: build reverse geocode request", locationsearch.ErrProvider)
	}
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set(
		"X-Goog-FieldMask",
		"results.placeId,results.formattedAddress,results.location",
	)

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return locationsearch.Place{}, err
		}
		return locationsearch.Place{}, fmt.Errorf("%w: reverse geocode request", locationsearch.ErrProvider)
	}
	defer resp.Body.Close()
	if err := providerStatus(resp.StatusCode); err != nil {
		return locationsearch.Place{}, err
	}

	var payload struct {
		Results []struct {
			PlaceID          string `json:"placeId"`
			FormattedAddress string `json:"formattedAddress"`
			Location         struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
		} `json:"results"`
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return locationsearch.Place{}, err
	}
	if len(payload.Results) == 0 {
		return locationsearch.Place{}, locationsearch.ErrNotFound
	}
	first := payload.Results[0]
	place := locationsearch.Place{
		PlaceID:     strings.TrimSpace(first.PlaceID),
		Label:       strings.TrimSpace(first.FormattedAddress),
		Location:    point,
		SnapToPlace: false,
	}
	if place.Label == "" {
		return locationsearch.Place{}, locationsearch.ErrNotFound
	}
	return place, nil
}

func suitableNamedPlaceCandidate(
	primaryType string,
	types []string,
	pureServiceAreaBusiness bool,
	businessStatus string,
) bool {
	if pureServiceAreaBusiness || strings.EqualFold(strings.TrimSpace(businessStatus), "CLOSED_PERMANENTLY") {
		return false
	}
	genericAddressTypes := map[string]struct{}{
		"geocode":        {},
		"intersection":   {},
		"plus_code":      {},
		"premise":        {},
		"route":          {},
		"street_address": {},
		"subpremise":     {},
	}
	primaryType = strings.TrimSpace(primaryType)
	if _, generic := genericAddressTypes[primaryType]; generic {
		return false
	}
	if primaryType != "" {
		return true
	}
	for _, placeType := range types {
		switch strings.TrimSpace(placeType) {
		case "landmark", "point_of_interest":
			return true
		}
	}
	return false
}

func distanceMeters(a, b locationsearch.Point) float64 {
	const earthRadiusMeters = 6371000.0
	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	deltaLat := (b.Latitude - a.Latitude) * math.Pi / 180
	deltaLng := (b.Longitude - a.Longitude) * math.Pi / 180

	h := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return 2 * earthRadiusMeters * math.Asin(math.Sqrt(h))
}

func providerStatus(status int) error {
	switch {
	case status >= 200 && status < 300:
		return nil
	case status == http.StatusBadRequest:
		return locationsearch.ErrInvalidInput
	case status == http.StatusNotFound:
		return locationsearch.ErrNotFound
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return locationsearch.ErrUnavailable
	default:
		return locationsearch.ErrProvider
	}
}

func decodeJSON(reader io.Reader, target any) error {
	if err := json.NewDecoder(reader).Decode(target); err != nil {
		return fmt.Errorf("%w: decode provider response", locationsearch.ErrProvider)
	}
	return nil
}
