package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type createRideRequestBody struct {
	Pickup struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"pickup"`
	Destination struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"destination"`
}

func (app application) createRideRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireRiderCapability(w, r)
	if !ok {
		return
	}

	var body createRideRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	request, err := app.rides.Create(r.Context(), u.ID,
		ride.Location{Latitude: body.Pickup.Latitude, Longitude: body.Pickup.Longitude},
		ride.Location{Latitude: body.Destination.Latitude, Longitude: body.Destination.Longitude},
	)
	if errors.Is(err, ride.ErrInvalidLocation) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup and destination coordinates are invalid"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to create ride request"})
		return
	}
	writeRideRequest(w, http.StatusCreated, request)
}

func (app application) requireRiderCapability(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	u, ok := app.currentUser(w, r)
	if !ok {
		return user.User{}, false
	}
	for _, capability := range u.Capabilities {
		if capability == user.CapabilityRider {
			return u, true
		}
	}
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "rider capability required"})
	return user.User{}, false
}

func writeRideRequest(w http.ResponseWriter, status int, request ride.Request) {
	writeJSON(w, status, map[string]any{
		"id":            request.ID,
		"rider_user_id": request.RiderUserID,
		"pickup": map[string]any{
			"latitude":  request.Pickup.Latitude,
			"longitude": request.Pickup.Longitude,
		},
		"destination": map[string]any{
			"latitude":  request.Destination.Latitude,
			"longitude": request.Destination.Longitude,
		},
		"status":     request.Status,
		"created_at": request.CreatedAt,
	})
}
