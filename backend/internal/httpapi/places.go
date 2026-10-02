package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

type placeAutocompleteRequest struct {
	Input        string      `json:"input"`
	SessionToken string      `json:"session_token"`
	Bias         *placePoint `json:"bias,omitempty"`
}

type placePoint struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type reverseGeocodeRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (api *API) autocompletePlaces(w http.ResponseWriter, r *http.Request) {
	user, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	if !api.allowLocationSearch(user.ID.String(), w) {
		return
	}

	var body placeAutocompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid autocomplete request"})
		return
	}
	input := locationsearch.AutocompleteInput{
		Query:        body.Input,
		SessionToken: body.SessionToken,
	}
	if body.Bias != nil && body.Bias.Latitude != nil && body.Bias.Longitude != nil {
		input.Bias = &locationsearch.Point{
			Latitude:  *body.Bias.Latitude,
			Longitude: *body.Bias.Longitude,
		}
	}

	suggestions, err := api.locationSearch.Autocomplete(r.Context(), input)
	if err != nil {
		api.writeLocationSearchError(w, r, err)
		return
	}
	items := make([]map[string]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		items = append(items, map[string]string{
			"place_id": suggestion.PlaceID,
			"label":    suggestion.Label,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestions": items})
}

func (api *API) getPlaceDetails(w http.ResponseWriter, r *http.Request) {
	user, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	if !api.allowLocationSearch(user.ID.String(), w) {
		return
	}
	place, err := api.locationSearch.Details(
		r.Context(),
		r.PathValue("place_id"),
		r.URL.Query().Get("session_token"),
	)
	if err != nil {
		api.writeLocationSearchError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, placeResponse(place))
}

func (api *API) reverseGeocodePlace(w http.ResponseWriter, r *http.Request) {
	user, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	if !api.allowLocationSearch(user.ID.String(), w) {
		return
	}

	var body reverseGeocodeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.Latitude == nil ||
		body.Longitude == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "latitude and longitude are required"})
		return
	}
	place, err := api.locationSearch.ReverseGeocode(r.Context(), locationsearch.Point{
		Latitude:  *body.Latitude,
		Longitude: *body.Longitude,
	})
	if err != nil {
		api.writeLocationSearchError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, placeResponse(place))
}

func (api *API) allowLocationSearch(userID string, w http.ResponseWriter) bool {
	if api.locationSearch == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "location search is unavailable"})
		return false
	}
	if api.locationSearchLimiter != nil &&
		!api.locationSearchLimiter.allow(userID, time.Now()) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "location search rate limit exceeded"})
		return false
	}
	return true
}

func (api *API) writeLocationSearchError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, context.Canceled) && r.Context().Err() != nil {
		return
	}
	status := locationSearchStatus(err)
	message := "location provider request failed"
	switch status {
	case http.StatusBadRequest:
		message = "invalid location search request"
	case http.StatusNotFound:
		message = "location was not found"
	case http.StatusServiceUnavailable:
		message = "location search is unavailable"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func locationSearchStatus(err error) int {
	switch {
	case errors.Is(err, locationsearch.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, locationsearch.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, locationsearch.ErrUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadGateway
	}
}

func placeResponse(place locationsearch.Place) map[string]any {
	response := map[string]any{
		"label":         strings.TrimSpace(place.Label),
		"latitude":      place.Location.Latitude,
		"longitude":     place.Location.Longitude,
	}
	if strings.TrimSpace(place.PlaceID) != "" {
		response["place_id"] = strings.TrimSpace(place.PlaceID)
	}
	return response
}
