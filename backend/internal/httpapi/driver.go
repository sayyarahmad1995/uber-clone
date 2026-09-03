package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
)

type onboardDriverRequest struct {
	DisplayName string `json:"display_name"`
	Vehicle     struct {
		Make         string `json:"make"`
		Model        string `json:"model"`
		ModelYear    int    `json:"model_year"`
		Color        string `json:"color"`
		LicensePlate string `json:"license_plate"`
	} `json:"vehicle"`
}

func (r onboardDriverRequest) input() driver.OnboardingInput {
	return driver.OnboardingInput{
		DisplayName: r.DisplayName,
		Vehicle: driver.VehicleInput{
			Make:         r.Vehicle.Make,
			Model:        r.Vehicle.Model,
			ModelYear:    r.Vehicle.ModelYear,
			Color:        r.Vehicle.Color,
			LicensePlate: r.Vehicle.LicensePlate,
		},
	}
}

type driverAvailabilityRequest struct {
	IsOnline bool `json:"is_online"`
}

func (api *API) onboardDriver(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	var body onboardDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	profile, err := api.drivers.Onboard(r.Context(), u.ID, body.input())
	if errors.Is(err, driver.ErrInvalidProfile) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "display_name and complete vehicle details including model_year are required"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to onboard driver"})
		return
	}
	writeDriver(w, http.StatusOK, profile)
}

func (api *API) getDriver(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	profile, err := api.drivers.Get(r.Context(), u.ID)
	if errors.Is(err, driver.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "driver profile not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load driver"})
		return
	}
	writeDriver(w, http.StatusOK, profile)
}

func (api *API) setDriverAvailability(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	var body driverAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	profile, err := api.drivers.SetOnline(r.Context(), u.ID, body.IsOnline)
	if errors.Is(err, driver.ErrNotFound) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "driver onboarding is required before changing availability"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update driver availability"})
		return
	}
	writeDriver(w, http.StatusOK, profile)
}

func (api *API) getDriverCurrentTrip(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	view, err := api.driverTrips.GetCurrent(r.Context(), u.ID)
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

func (api *API) listDriverTripHistory(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	views, err := api.driverTrips.ListHistory(r.Context(), u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to list trip history"})
		return
	}
	writeJSON(w, http.StatusOK, driverTripHistoryResponse(views))
}

func writeDriver(w http.ResponseWriter, status int, profile driver.Profile) {
	writeJSON(w, status, map[string]any{
		"user_id":      profile.UserID,
		"display_name": profile.DisplayName,
		"status":       profile.Status,
		"is_online":    profile.IsOnline,
		"vehicle": map[string]any{
			"id":            profile.Vehicle.ID,
			"make":          profile.Vehicle.Make,
			"model":         profile.Vehicle.Model,
			"model_year":    modelYearResponse(profile.Vehicle.ModelYear),
			"color":         profile.Vehicle.Color,
			"license_plate": profile.Vehicle.LicensePlate,
		},
	})
}

// Zero is the internal legacy sentinel, not a real vehicle model year.
func modelYearResponse(year int) any {
	if year == 0 {
		return nil
	}
	return year
}

func driverCurrentTripResponse(view drivertrip.View) map[string]any {
	return map[string]any{
		"ride_request_id": view.RideRequestID,
		"pickup":          map[string]any{"latitude": view.Pickup.Latitude, "longitude": view.Pickup.Longitude},
		"destination":     map[string]any{"latitude": view.Destination.Latitude, "longitude": view.Destination.Longitude},
		"status":          view.Status,
		"assigned_at":     view.AssignedAt,
		"started_at":      view.StartedAt,
	}
}

func driverTripHistoryResponse(views []drivertrip.View) map[string]any {
	trips := make([]map[string]any, 0, len(views))
	for _, view := range views {
		trips = append(trips, map[string]any{
			"ride_request_id": view.RideRequestID,
			"pickup":          map[string]any{"latitude": view.Pickup.Latitude, "longitude": view.Pickup.Longitude},
			"destination":     map[string]any{"latitude": view.Destination.Latitude, "longitude": view.Destination.Longitude},
			"status":          view.Status,
			"assigned_at":     view.AssignedAt,
			"started_at":      view.StartedAt,
			"completed_at":    view.CompletedAt,
			"cancelled_at":    view.CancelledAt,
		})
	}
	return map[string]any{"trips": trips}
}
