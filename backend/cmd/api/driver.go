package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type onboardDriverRequest struct {
	Vehicle struct {
		Make         string `json:"make"`
		Model        string `json:"model"`
		Color        string `json:"color"`
		LicensePlate string `json:"license_plate"`
	} `json:"vehicle"`
}

func (r onboardDriverRequest) vehicleInput() driver.VehicleInput {
	return driver.VehicleInput{
		Make:         r.Vehicle.Make,
		Model:        r.Vehicle.Model,
		Color:        r.Vehicle.Color,
		LicensePlate: r.Vehicle.LicensePlate,
	}
}

type driverAvailabilityRequest struct {
	IsOnline bool `json:"is_online"`
}

func (app application) onboardDriver(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}
	var body onboardDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	profile, err := app.drivers.Onboard(r.Context(), u.ID, body.vehicleInput())
	if errors.Is(err, driver.ErrInvalidProfile) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "vehicle make, model, color, and license_plate are required"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to onboard driver"})
		return
	}
	writeDriver(w, http.StatusOK, profile)
}

func (app application) getDriver(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}
	profile, err := app.drivers.Get(r.Context(), u.ID)
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

func (app application) setDriverAvailability(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}
	var body driverAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	profile, err := app.drivers.SetOnline(r.Context(), u.ID, body.IsOnline)
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

func (app application) requireDriverCapability(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	u, ok := app.currentUser(w, r)
	if !ok {
		return user.User{}, false
	}
	for _, capability := range u.Capabilities {
		if capability == user.CapabilityDriver {
			return u, true
		}
	}
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "driver capability required"})
	return user.User{}, false
}

func writeDriver(w http.ResponseWriter, status int, profile driver.Profile) {
	writeJSON(w, status, map[string]any{
		"user_id":   profile.UserID,
		"status":    profile.Status,
		"is_online": profile.IsOnline,
		"vehicle": map[string]any{
			"id":            profile.Vehicle.ID,
			"make":          profile.Vehicle.Make,
			"model":         profile.Vehicle.Model,
			"color":         profile.Vehicle.Color,
			"license_plate": profile.Vehicle.LicensePlate,
		},
	})
}
