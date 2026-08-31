package main

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/matching"
)

func (app application) matchRideRequest(w http.ResponseWriter,r *http.Request) {
	u,ok := app.requireRiderCapability(w,r); if !ok { return }
	rideRequestID,err := uuid.Parse(r.PathValue("ride_request_id")); if err != nil { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid ride_request_id"}); return }
	result,err := app.matching.Match(r.Context(),rideRequestID,u.ID)
	switch {
	case errors.Is(err,matching.ErrRideNotFound): writeJSON(w,http.StatusNotFound,map[string]string{"error":"ride request not found"}); return
	case errors.Is(err,matching.ErrRideNotMatchable): writeJSON(w,http.StatusConflict,map[string]string{"error":"ride request uses the offers marketplace"}); return
	case errors.Is(err,matching.ErrNoEligibleDriver): writeJSON(w,http.StatusConflict,map[string]string{"error":"no eligible driver available"}); return
	case err != nil: writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"unable to match ride request"}); return
	}
	status := http.StatusOK; if result.Created { status = http.StatusCreated }
	writeJSON(w,status,map[string]any{"ride_request_id":result.Candidate.RideRequestID,"driver_user_id":result.Candidate.DriverUserID,"created_at":result.Candidate.CreatedAt})
}

func (app application) rejectRideRequestCandidate(w http.ResponseWriter,r *http.Request) {
	u,ok := app.requireDriverCapability(w,r); if !ok { return }
	rideRequestID,err := uuid.Parse(r.PathValue("ride_request_id")); if err != nil { writeJSON(w,http.StatusBadRequest,map[string]string{"error":"invalid ride_request_id"}); return }
	candidate,err := app.matching.Reject(r.Context(),rideRequestID,u.ID)
	switch {
	case errors.Is(err,matching.ErrCandidateNotFound): writeJSON(w,http.StatusNotFound,map[string]string{"error":"ride request candidate not found"}); return
	case errors.Is(err,matching.ErrCandidateResolved): writeJSON(w,http.StatusConflict,map[string]string{"error":"ride request candidate already resolved"}); return
	case err != nil: writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"unable to update ride request candidate"}); return
	}
	writeCandidateDecision(w,candidate)
}

func writeCandidateDecision(w http.ResponseWriter,candidate matching.Candidate) { writeJSON(w,http.StatusOK,map[string]any{"ride_request_id":candidate.RideRequestID,"driver_user_id":candidate.DriverUserID,"status":candidate.Status,"created_at":candidate.CreatedAt,"decided_at":candidate.DecidedAt}) }
