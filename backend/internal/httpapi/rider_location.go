package httpapi

import (
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/riderlocation"
)

func (api *API) getRiderActiveTripDriverLocation(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}

	view, err := api.riderLocations.GetForActiveTrip(r.Context(), u.ID)
	switch {
	case errors.Is(err, riderlocation.ErrActiveTripNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "active trip not found"})
		return
	case errors.Is(err, riderlocation.ErrLocationNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "driver location not available"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to get driver location"})
		return
	}

	writeJSON(w, http.StatusOK, riderActiveTripDriverLocationResponse(view))
}

func riderActiveTripDriverLocationResponse(view riderlocation.View) map[string]any {
	return map[string]any{
		"ride_request_id": view.RideRequestID,
		"latitude":        view.Latitude,
		"longitude":       view.Longitude,
		"updated_at":      view.UpdatedAt,
	}
}
