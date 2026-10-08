package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

const (
	defaultPlacesBase       = "https://places.googleapis.com"
	defaultGeocodeBase      = "https://geocode.googleapis.com"
	popularityCacheLifetime = 24 * time.Hour
)

type Provider struct {
	apiKey      string
	client      *http.Client
	placesBase  string
	geocodeBase string

	popularityMu    sync.Mutex
	popularityCache map[string]cachedPopularity
}

type cachedPopularity struct {
	userRatingCount int64
	expiresAt       time.Time
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
		apiKey:          strings.TrimSpace(apiKey),
		client:          client,
		placesBase:      strings.TrimRight(placesBase, "/"),
		geocodeBase:     strings.TrimRight(geocodeBase, "/"),
		popularityCache: make(map[string]cachedPopularity),
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
		body["locationRestriction"] = map[string]any{
			"circle": map[string]any{
				"center": map[string]any{
					"latitude":  input.Bias.Latitude,
					"longitude": input.Bias.Longitude,
				},
				"radius": float64(input.RestrictionRadiusMeters),
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
	return p.rankSuggestionsByPopularity(ctx, items), nil
}

func (p *Provider) rankSuggestionsByPopularity(
	ctx context.Context,
	items []locationsearch.Suggestion,
) []locationsearch.Suggestion {
	if len(items) < 2 {
		return items
	}

	type rankedSuggestion struct {
		suggestion locationsearch.Suggestion
		popularity int64
		index      int
	}
	ranked := make([]rankedSuggestion, len(items))
	var wait sync.WaitGroup
	for i, item := range items {
		ranked[i] = rankedSuggestion{suggestion: item, index: i}
		wait.Add(1)
		go func(index int, placeID string) {
			defer wait.Done()
			ranked[index].popularity = p.userRatingCount(ctx, placeID)
		}(i, item.PlaceID)
	}
	wait.Wait()

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].popularity == ranked[j].popularity {
			return ranked[i].index < ranked[j].index
		}
		return ranked[i].popularity > ranked[j].popularity
	})
	for i := range ranked {
		items[i] = ranked[i].suggestion
	}
	return items
}

func (p *Provider) userRatingCount(ctx context.Context, placeID string) int64 {
	now := time.Now()
	p.popularityMu.Lock()
	if cached, ok := p.popularityCache[placeID]; ok && now.Before(cached.expiresAt) {
		p.popularityMu.Unlock()
		return cached.userRatingCount
	}
	p.popularityMu.Unlock()

	endpoint, err := url.Parse(p.placesBase + "/v1/places/" + url.PathEscape(placeID))
	if err != nil {
		return 0
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0
	}
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set("X-Goog-FieldMask", "userRatingCount")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0
	}

	var payload struct {
		UserRatingCount int64 `json:"userRatingCount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0
	}

	p.popularityMu.Lock()
	p.popularityCache[placeID] = cachedPopularity{
		userRatingCount: payload.UserRatingCount,
		expiresAt:       now.Add(popularityCacheLifetime),
	}
	p.popularityMu.Unlock()
	return payload.UserRatingCount
}

func (p *Provider) Details(ctx context.Context, placeID, sessionToken string) (locationsearch.Place, error) {
	if p.apiKey == "" {
		return locationsearch.Place{}, locationsearch.ErrUnavailable
	}
	endpoint, err := url.Parse(p.placesBase + "/v1/places/" + url.PathEscape(placeID))
	if err != nil {
		return locationsearch.Place{}, fmt.Errorf("%w: build place details URL", locationsearch.ErrProvider)
	}
	if sessionToken = strings.TrimSpace(sessionToken); sessionToken != "" {
		query := endpoint.Query()
		query.Set("sessionToken", sessionToken)
		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return locationsearch.Place{}, fmt.Errorf("%w: build place details request", locationsearch.ErrProvider)
	}
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set(
		"X-Goog-FieldMask",
		"id,displayName.text,formattedAddress,location",
	)

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
		ID          string `json:"id"`
		DisplayName struct {
			Text string `json:"text"`
		} `json:"displayName"`
		FormattedAddress string `json:"formattedAddress"`
		Location         struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}
	if err := decodeJSON(resp.Body, &payload); err != nil {
		return locationsearch.Place{}, err
	}
	label := strings.TrimSpace(payload.FormattedAddress)
	if strings.TrimSpace(sessionToken) == "" {
		if displayName := strings.TrimSpace(payload.DisplayName.Text); displayName != "" {
			label = displayName
		}
	}
	place := locationsearch.Place{
		PlaceID: strings.TrimSpace(payload.ID),
		Label:   label,
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

func (p *Provider) ReverseGeocode(ctx context.Context, point locationsearch.Point) (locationsearch.Place, error) {
	if p.apiKey == "" {
		return locationsearch.Place{}, locationsearch.ErrUnavailable
	}
	return p.reverseGeocodeAddress(ctx, point)
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
		PlaceID:  strings.TrimSpace(first.PlaceID),
		Label:    strings.TrimSpace(first.FormattedAddress),
		Location: point,
	}
	if place.Label == "" {
		return locationsearch.Place{}, locationsearch.ErrNotFound
	}
	return place, nil
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
