package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
)

type ridePreviewRequest struct {
	Pickup      routePoint `json:"pickup"`
	Destination routePoint `json:"destination"`
	ServiceCode string     `json:"service_code"`
}

type routePoint struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	PlaceID   string   `json:"place_id,omitempty"`
}

func (p routePoint) endpoint() routing.Endpoint {
	return routing.Endpoint{
		Point: routing.Point{
			Latitude:  *p.Latitude,
			Longitude: *p.Longitude,
		},
		PlaceID: strings.TrimSpace(p.PlaceID),
	}
}

func (api *API) createRidePreview(w http.ResponseWriter, r *http.Request) {
	if _, ok := api.requireRiderCapability(w, r); !ok {
		return
	}
	if api.routing == nil || api.pricing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "ride preview is unavailable"})
		return
	}

	var body ridePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.Pickup.Latitude == nil ||
		body.Pickup.Longitude == nil ||
		body.Destination.Latitude == nil ||
		body.Destination.Longitude == nil ||
		strings.TrimSpace(body.ServiceCode) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup, destination, and service_code are required"})
		return
	}

	route, err := api.routing.Preview(
		r.Context(),
		body.Pickup.endpoint(),
		body.Destination.endpoint(),
	)
	if err != nil {
		api.writeRoutingError(w, r, err)
		return
	}

	suggestion, err := api.pricing.Suggest(
		r.Context(),
		body.ServiceCode,
		route.DistanceMeters,
		route.DurationSeconds,
	)
	if err != nil {
		api.writePricingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ridePreviewResponse(route, suggestion))
}

func (api *API) writeRoutingError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, context.Canceled) && r.Context().Err() != nil {
		return
	}
	status := routingStatus(err)
	message := "route provider request failed"
	switch status {
	case http.StatusBadRequest:
		message = "invalid route preview request"
	case http.StatusNotFound:
		message = "no driving route was found"
	case http.StatusServiceUnavailable:
		message = "route preview is unavailable"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

func (api *API) writePricingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pricing.ErrPolicyUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "suggested fare is unavailable for this service",
		})
	case errors.Is(err, pricing.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid pricing request",
		})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": "suggested fare calculation failed",
		})
	}
}

func routingStatus(err error) int {
	switch {
	case errors.Is(err, routing.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, routing.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, routing.ErrUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadGateway
	}
}

func ridePreviewResponse(
	route routing.Route,
	suggestion pricing.Suggestion,
) map[string]any {
	return map[string]any{
		"route": map[string]any{
			"distance_meters":  route.DistanceMeters,
			"duration_seconds": route.DurationSeconds,
			"encoded_polyline": route.EncodedPolyline,
		},
		"suggested_fare": map[string]any{
			"amount_minor": suggestion.Fare.AmountMinor,
			"currency":     suggestion.Fare.Currency,
		},
		"pricing_policy_version": suggestion.PolicyVersion,
	}
}
