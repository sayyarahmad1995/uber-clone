package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/matching"
)

func TestWriteCandidateDecisionUsesApplicationOwnedContract(t *testing.T) {
	decidedAt := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	candidate := matching.Candidate{
		RideRequestID: uuid.New(),
		DriverUserID:  uuid.New(),
		Status:        matching.CandidateStatusAccepted,
		CreatedAt:     decidedAt.Add(-time.Minute),
		DecidedAt:     &decidedAt,
	}

	recorder := httptest.NewRecorder()
	writeCandidateDecision(recorder, candidate)

	if recorder.Code != 200 {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var payload struct {
		RideRequestID uuid.UUID                `json:"ride_request_id"`
		DriverUserID  uuid.UUID                `json:"driver_user_id"`
		Status        matching.CandidateStatus `json:"status"`
		CreatedAt     time.Time                `json:"created_at"`
		DecidedAt     *time.Time               `json:"decided_at"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.RideRequestID != candidate.RideRequestID || payload.DriverUserID != candidate.DriverUserID {
		t.Fatalf("unexpected candidate ids: %#v", payload)
	}
	if payload.Status != matching.CandidateStatusAccepted || payload.DecidedAt == nil || !payload.DecidedAt.Equal(decidedAt) {
		t.Fatalf("unexpected decision payload: %#v", payload)
	}
}
