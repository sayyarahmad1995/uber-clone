package offer

import (
	"context"
	"errors"
	"testing"
)

func TestRiderReadRejectsExpiredDeadlineBeforeExpirySweep(t *testing.T) {
	db := openOfferIntegrationDB(t)
	riderID := createOfferTestUser(t, db, "rider")
	rideID := createOfferTestRide(t, db, riderID)

	if _, err := db.Exec(`
		UPDATE ride_requests
		SET expires_at=statement_timestamp()-INTERVAL '1 second'
		WHERE id=$1
	`, rideID); err != nil {
		t.Fatal(err)
	}

	var before string
	if err := db.QueryRow(`SELECT status FROM ride_requests WHERE id=$1`, rideID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if before != "requested" {
		t.Fatalf("fixture status=%q want stale requested", before)
	}

	_, err := NewPostgresRepository(db).ListForRider(context.Background(), rideID, riderID)
	if !errors.Is(err, ErrRideNotOpen) {
		t.Fatalf("ListForRider error=%v want ErrRideNotOpen", err)
	}

	var after string
	if err := db.QueryRow(`SELECT status FROM ride_requests WHERE id=$1`, rideID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != "requested" {
		t.Fatalf("business read materialized status=%q; periodic sweep must own materialization", after)
	}
}
