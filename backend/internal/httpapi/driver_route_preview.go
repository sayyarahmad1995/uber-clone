package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type driverRouteRequest struct {
	RideRequestID string     `json:"ride_request_id"`
	Status        string     `json:"status"`
	Origin        routePoint `json:"origin"`
}

func (api *API) createDriverRoutePreview(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	var body driverRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid route request"})
		return
	}
	requestID, err := uuid.Parse(body.RideRequestID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}
	// Ownership and trip endpoints come from the authenticated Driver's view.
	view, err := api.driverTrips.GetCurrent(r.Context(), u.ID)
	switch {
	case errors.Is(err, drivertrip.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "active trip not found"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to get current trip"})
		return
	}
	if view.RideRequestID != requestID || string(view.Status) != body.Status {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "trip or route stage changed"})
		return
	}
	pickup := routing.Endpoint{Point: routing.Point{
		Latitude: view.Pickup.Latitude, Longitude: view.Pickup.Longitude,
	}}
	destination := routing.Endpoint{Point: routing.Point{
		Latitude: view.Destination.Latitude, Longitude: view.Destination.Longitude,
	}}
	var from, to routing.Endpoint
	switch view.Status {
	case trip.StatusAssigned:
		if body.Origin.Latitude == nil || body.Origin.Longitude == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "current Driver coordinates are required"})
			return
		}
		from = routing.Endpoint{Point: body.Origin.endpoint().Point}
		to = pickup
	case trip.StatusInProgress:
		from, to = pickup, destination
	default:
		writeJSON(w, http.StatusConflict, map[string]string{"error": "trip has no active driving route"})
		return
	}
	if api.routing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "route preview is unavailable"})
		return
	}
	route, err := api.routing.Preview(r.Context(), from, to)
	if err != nil {
		api.writeRoutingError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, ridePreviewResponse(route))
}
