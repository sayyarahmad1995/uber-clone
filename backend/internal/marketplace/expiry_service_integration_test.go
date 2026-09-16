package marketplace

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func TestExpiryServiceSweepMaterializesMarketplaceTerminalStates(t *testing.T) {
	db := openTripIntegrationDB(t)
	ctx := context.Background()
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)

	expiredOfferRide := createTripIntegrationRide(t, db, riderID)
	insertTripIntegrationOffer(t, db, expiredOfferRide, driverID)
	if _, err := db.Exec(`
		UPDATE ride_offers
		SET created_at=statement_timestamp()-INTERVAL '2 minutes',
		    updated_at=statement_timestamp()-INTERVAL '2 minutes',
		    expires_at=statement_timestamp()-INTERVAL '1 second'
		WHERE ride_request_id=$1
	`, expiredOfferRide); err != nil {
		t.Fatal(err)
	}

	expiredWindowRide := createTripIntegrationRide(t, db, riderID)
	windowDriverID := createTripIntegrationDriver(t, db)
	if _, err := db.Exec(`
		INSERT INTO driver_ride_request_opportunities (ride_request_id,driver_user_id,status,visible_until)
		VALUES ($1,$2,'open',statement_timestamp()-INTERVAL '1 second')
	`, expiredWindowRide, windowDriverID); err != nil {
		t.Fatal(err)
	}

	expiredRide := createTripIntegrationRide(t, db, riderID)
	if _, err := db.Exec(`
		UPDATE ride_requests
		SET created_at=statement_timestamp()-INTERVAL '2 minutes',
		    expires_at=statement_timestamp()-INTERVAL '1 second'
		WHERE id=$1
	`, expiredRide); err != nil {
		t.Fatal(err)
	}

	closedRides := make(map[string]uuid.UUID)
	for _, status := range []string{"accepted", "cancelled"} {
		rideID := createTripIntegrationRide(t, db, riderID)
		closureDriverID := createTripIntegrationDriver(t, db)
		insertTripIntegrationOffer(t, db, rideID, closureDriverID)
		if _, err := db.Exec(`UPDATE ride_requests SET status=$2 WHERE id=$1`, rideID, status); err != nil {
			t.Fatal(err)
		}
		closedRides[status] = rideID
	}

	service := NewExpiryService(db)
	if err := service.Sweep(ctx); err != nil {
		t.Fatalf("Sweep returned error: %v", err)
	}

	assertMarketplaceStatus(t, db, "ride_offers", expiredOfferRide, "expired")
	assertMarketplaceStatus(t, db, "driver_ride_request_opportunities", expiredOfferRide, "offer_expired")
	assertMarketplaceStatus(t, db, "driver_ride_request_opportunities", expiredWindowRide, "window_expired")
	assertMarketplaceStatus(t, db, "ride_requests", expiredRide, "expired")

	for terminal, rideID := range closedRides {
		t.Run("closure_after_"+terminal, func(t *testing.T) {
			assertMarketplaceStatus(t, db, "ride_offers", rideID, "closed")
			assertMarketplaceStatus(t, db, "driver_ride_request_opportunities", rideID, "closed")
		})
	}
}

func assertMarketplaceStatus(t *testing.T, db interface {
	QueryRow(query string, args ...any) *sql.Row
}, table string, rideID uuid.UUID, want string) {
	t.Helper()
	var got string
	query := "SELECT status FROM " + table + " WHERE "
	if table == "ride_requests" {
		query += "id=$1"
	} else {
		query += "ride_request_id=$1"
	}
	if err := db.QueryRow(query, rideID).Scan(&got); err != nil {
		t.Fatalf("read %s status: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s status=%q want %q", table, got, want)
	}
}
