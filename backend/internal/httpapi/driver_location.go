package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
)

type driverLocationRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (api *API) setDriverLocation(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}

	var body driverLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Latitude == nil || body.Longitude == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "latitude and longitude are required"})
		return
	}

	location, err := api.driverLocations.SetCurrent(r.Context(), u.ID, driverlocation.Input{
		Latitude:  *body.Latitude,
		Longitude: *body.Longitude,
	})
	switch {
	case errors.Is(err, driverlocation.ErrInvalidLocation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "latitude or longitude is out of range"})
		return
	case errors.Is(err, driverlocation.ErrDriverNotFound):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "driver onboarding is required before updating location"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update driver location"})
		return
	}

	writeJSON(w, http.StatusOK, driverLocationResponse(location))
}

func driverLocationResponse(location driverlocation.Location) map[string]any {
	return map[string]any{
		"latitude":   location.Latitude,
		"longitude":  location.Longitude,
		"updated_at": location.UpdatedAt,
	}
}
