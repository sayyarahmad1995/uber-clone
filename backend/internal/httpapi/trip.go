package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type tripTransitionFunc func(context.Context, uuid.UUID, uuid.UUID) (trip.Trip, error)

func (api *API) acceptRideRequestCandidate(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}
	acceptance, err := api.trips.Accept(r.Context(), rideRequestID, u.ID)
	if errors.Is(err, trip.ErrAssignmentNotFound) {
		marketplaceAcceptance, marketplaceErr := api.offers.AcceptProposed(r.Context(), rideRequestID, u.ID)
		if marketplaceErr == nil {
			response := rideOfferResponse(marketplaceAcceptance.Offer)
			response["trip"] = tripResponse(*marketplaceAcceptance.Trip)
			writeJSON(w, http.StatusOK, response)
			return
		}
		if errors.Is(marketplaceErr, offer.ErrRideNotOpen) {
			writeOfferError(w, marketplaceErr)
			return
		}
		if !errors.Is(marketplaceErr, offer.ErrRideNotFound) {
			if writeOfferError(w, marketplaceErr) {
				return
			}
		}
	}
	switch {
	case errors.Is(err, trip.ErrAssignmentNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ride request candidate not found"})
		return
	case errors.Is(err, trip.ErrAssignmentResolved):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "ride request candidate already resolved"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update ride request candidate"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ride_request_id": acceptance.Trip.RideRequestID,
		"driver_user_id":  acceptance.Trip.DriverUserID,
		"status":          "accepted",
		"created_at":      acceptance.CandidateCreatedAt,
		"decided_at":      acceptance.CandidateDecidedAt,
	})
}

func (api *API) startTrip(w http.ResponseWriter, r *http.Request) {
	api.transitionTrip(w, r, api.trips.Start)
}

func (api *API) completeTrip(w http.ResponseWriter, r *http.Request) {
	api.transitionTrip(w, r, api.trips.Complete)
}

func (api *API) transitionTrip(w http.ResponseWriter, r *http.Request, transition tripTransitionFunc) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}
	result, err := transition(r.Context(), rideRequestID, u.ID)
	switch {
	case errors.Is(err, trip.ErrTripNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trip not found"})
		return
	case errors.Is(err, trip.ErrTripNotStarted):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "trip has not started"})
		return
	case errors.Is(err, trip.ErrTripCompleted):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "trip already completed"})
		return
	case errors.Is(err, trip.ErrTripCancelled):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "trip already cancelled"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update trip"})
		return
	}
	writeTrip(w, result)
}

func writeTrip(w http.ResponseWriter, result trip.Trip) {
	writeJSON(w, http.StatusOK, tripResponse(result))
}

func tripResponse(result trip.Trip) map[string]any {
	return map[string]any{
		"ride_request_id": result.RideRequestID,
		"rider_user_id":   result.RiderUserID,
		"driver_user_id":  result.DriverUserID,
		"status":          result.Status,
		"assigned_at":     result.AssignedAt,
		"started_at":      result.StartedAt,
		"completed_at":    result.CompletedAt,
		"cancelled_at":    result.CancelledAt,
	}
}
