package matching

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestPostgresRepositoryRejectExpiredCandidateReleasesWithoutRejection(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	driverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC()))
	rideID := createIntegrationRide(t, db, riderID, "automatic")
	insertCandidate(t, db, rideID, driverID, time.Now().UTC().Add(-45*time.Second), nil)

	_, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Reject(context.Background(), rideID, driverID)
	if !errors.Is(err, ErrCandidateResolved) {
		t.Fatalf("expected ErrCandidateResolved, got %v", err)
	}

	var status string
	var decidedAt, releasedAt sql.NullTime
	if err := db.QueryRow(`
		SELECT status, decided_at, released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, driverID).Scan(&status, &decidedAt, &releasedAt); err != nil {
		t.Fatalf("select candidate: %v", err)
	}
	if status != string(CandidateStatusPending) || decidedAt.Valid || !releasedAt.Valid {
		t.Fatalf("expected released pending candidate without rejection decision, status=%s decided=%v released=%v", status, decidedAt.Valid, releasedAt.Valid)
	}
}

func TestPostgresRepositoryRejectFreshCandidateStillRejects(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	driverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC()))
	rideID := createIntegrationRide(t, db, riderID, "automatic")
	insertCandidate(t, db, rideID, driverID, time.Now().UTC(), nil)

	candidate, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Reject(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("reject fresh candidate: %v", err)
	}
	if candidate.Status != CandidateStatusRejected || candidate.DecidedAt == nil {
		t.Fatalf("expected rejected candidate with decision timestamp, got %#v", candidate)
	}

	var releasedAt sql.NullTime
	if err := db.QueryRow(`
		SELECT released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, driverID).Scan(&releasedAt); err != nil {
		t.Fatalf("select candidate release: %v", err)
	}
	if releasedAt.Valid {
		t.Fatalf("fresh rejection should not use release marker, got %v", releasedAt.Time)
	}
}
