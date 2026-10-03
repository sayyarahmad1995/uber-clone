package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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
}

func (api *API) createRidePreview(w http.ResponseWriter, r *http.Request) {
	if _, ok := api.requireRiderCapability(w, r); !ok {
		return
	}
	if api.routing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "route preview is unavailable"})
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

	preview, err := api.routing.Preview(
		r.Context(),
		routing.Point{
			Latitude:  *body.Pickup.Latitude,
			Longitude: *body.Pickup.Longitude,
		},
		routing.Point{
			Latitude:  *body.Destination.Latitude,
			Longitude: *body.Destination.Longitude,
		},
	)
	if err != nil {
		api.writeRoutingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, ridePreviewResponse(preview))
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

func ridePreviewResponse(preview routing.Preview) map[string]any {
	routes := make([]map[string]any, 0, len(preview.Routes))
	for _, route := range preview.Routes {
		routes = append(routes, map[string]any{
			"id":               route.ID,
			"recommended":      route.Recommended,
			"distance_meters":  route.DistanceMeters,
			"duration_seconds": route.DurationSeconds,
			"encoded_polyline": route.EncodedPolyline,
		})
	}
	return map[string]any{"routes": routes}
}
