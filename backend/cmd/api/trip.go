package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type tripTransitionFunc func(context.Context, uuid.UUID, uuid.UUID) (trip.Trip, error)

func (app application) acceptRideRequestCandidate(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}

	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}

	acceptance, err := app.trips.Accept(r.Context(), rideRequestID, u.ID)
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

func (app application) startTrip(w http.ResponseWriter, r *http.Request) {
	app.transitionTrip(w, r, app.trips.Start)
}

func (app application) completeTrip(w http.ResponseWriter, r *http.Request) {
	app.transitionTrip(w, r, app.trips.Complete)
}

func (app application) transitionTrip(w http.ResponseWriter, r *http.Request, transition tripTransitionFunc) {
	u, ok := app.requireDriverCapability(w, r)
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
	}
}
