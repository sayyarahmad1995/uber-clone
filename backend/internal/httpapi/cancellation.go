package httpapi

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/cancellation"
)

func (api *API) cancelRideRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}
	result, err := api.cancellations.CancelByRider(r.Context(), rideRequestID, u.ID)
	if writeCancellationError(w, err) {
		return
	}
	writeCancellation(w, result)
}

func (api *API) cancelDriverRideRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}
	result, err := api.cancellations.CancelByDriver(r.Context(), rideRequestID, u.ID)
	if writeCancellationError(w, err) {
		return
	}
	writeCancellation(w, result)
}

func writeCancellationError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, cancellation.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ride request or trip not found"})
	case errors.Is(err, cancellation.ErrTripCompleted):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "completed trip cannot be cancelled"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to cancel ride"})
	}
	return true
}

func writeCancellation(w http.ResponseWriter, result cancellation.Result) {
	response := map[string]any{
		"ride_request_id": result.RideRequestID,
		"status":          result.Status,
		"cancelled_by":    result.CancelledBy,
		"cancelled_at":    result.CancelledAt,
		"trip":            nil,
	}
	if result.Trip != nil {
		response["trip"] = tripResponse(*result.Trip)
	}
	writeJSON(w, http.StatusOK, response)
}
