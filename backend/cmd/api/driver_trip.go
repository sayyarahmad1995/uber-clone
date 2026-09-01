package main

import (
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
)

func (app application) getDriverCurrentTrip(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}

	view, err := app.driverTrips.GetCurrent(r.Context(), u.ID)
	switch {
	case errors.Is(err, drivertrip.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "active trip not found"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to get current trip"})
		return
	}

	writeJSON(w, http.StatusOK, driverCurrentTripResponse(view))
}

func driverCurrentTripResponse(view drivertrip.View) map[string]any {
	return map[string]any{
		"ride_request_id": view.RideRequestID,
		"pickup": map[string]any{
			"latitude":  view.Pickup.Latitude,
			"longitude": view.Pickup.Longitude,
		},
		"destination": map[string]any{
			"latitude":  view.Destination.Latitude,
			"longitude": view.Destination.Longitude,
		},
		"status":      view.Status,
		"assigned_at": view.AssignedAt,
		"started_at":  view.StartedAt,
	}
}
