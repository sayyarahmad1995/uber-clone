package trip

import (
	"context"
	"testing"
)

func TestPostgresSelectionMarksRideRequestAccepted(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	insertTripIntegrationOffer(t, db, rideID, driverID)

	if _, err := NewPostgresRepository(db).SelectOffer(context.Background(), rideID, riderID, driverID); err != nil {
		t.Fatalf("select offer: %v", err)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM ride_requests WHERE id = $1`, rideID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "accepted" {
		t.Fatalf("expected accepted ride request, got %q", status)
	}
}

func TestPostgresSelectionBackfillMarksExistingTripsAccepted(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)

	if _, err := db.Exec(`
		INSERT INTO trips (ride_request_id, rider_user_id, driver_user_id, assigned_at)
		VALUES ($1, $2, $3, NOW())
	`, rideID, riderID, driverID); err != nil {
		t.Fatalf("insert legacy trip fixture: %v", err)
	}
	if _, err := db.Exec(`UPDATE ride_requests SET status = 'requested' WHERE id = $1`, rideID); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`
		UPDATE ride_requests rr
		SET status = 'accepted'
		WHERE status = 'requested'
		  AND EXISTS (
			  SELECT 1
			  FROM trips t
			  WHERE t.ride_request_id = rr.id
				AND t.status IN ('assigned', 'in_progress', 'completed')
		  )
	`); err != nil {
		t.Fatal(err)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM ride_requests WHERE id = $1`, rideID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "accepted" {
		t.Fatalf("expected accepted ride request after backfill, got %q", status)
	}
}
