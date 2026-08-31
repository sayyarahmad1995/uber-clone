package main

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func (app application) getRideRequestStatus(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireRiderCapability(w, r)
	if !ok {
		return
	}

	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}

	view, err := app.rideStatuses.GetOwned(r.Context(), rideRequestID, u.ID)
	switch {
	case errors.Is(err, ridestatus.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ride request not found"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to get ride request status"})
		return
	}

	writeJSON(w, http.StatusOK, rideRequestStatusResponse(view.RideRequest, view.Trip))
}

func rideRequestStatusResponse(request ride.Request, assignedTrip *trip.Trip) map[string]any {
	response := map[string]any{
		"id":           request.ID,
		"pickup":       map[string]any{"latitude": request.Pickup.Latitude, "longitude": request.Pickup.Longitude},
		"destination":  map[string]any{"latitude": request.Destination.Latitude, "longitude": request.Destination.Longitude},
		"booking_mode": request.BookingMode,
		"status":       request.Status,
		"created_at":   request.CreatedAt,
		"trip":         nil,
	}
	if request.ProposedFare != nil {
		response["proposed_fare"] = map[string]any{
			"amount_minor": request.ProposedFare.AmountMinor,
			"currency":     request.ProposedFare.Currency,
		}
	}
	if assignedTrip != nil {
		response["trip"] = map[string]any{
			"driver_user_id": assignedTrip.DriverUserID,
			"status":         assignedTrip.Status,
			"assigned_at":    assignedTrip.AssignedAt,
			"started_at":     assignedTrip.StartedAt,
			"completed_at":   assignedTrip.CompletedAt,
		}
	}
	return response
}
