package trip

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func openTripIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatalf("TEST_DATABASE_URL must point to a database ending in _test, got %q", parsed.Path)
	}
	db, err := database.Open(databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Apply(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func createTripIntegrationUser(t *testing.T, db *sql.DB, capability string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_capabilities (user_id, capability) VALUES ($1, $2)`, userID, capability); err != nil {
		t.Fatalf("insert capability: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
	return userID
}

func createTripIntegrationDriver(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := createTripIntegrationUser(t, db, "driver")
	if _, err := db.Exec(`INSERT INTO driver_profiles (user_id, status, is_online) VALUES ($1, 'active', TRUE)`, userID); err != nil {
		t.Fatalf("insert driver profile: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO driver_vehicles (id, driver_user_id, make, model, color, license_plate) VALUES ($1, $2, 'Test', 'Car', 'White', 'TEST')`, uuid.New(), userID); err != nil {
		t.Fatalf("insert vehicle: %v", err)
	}
	return userID
}

func createTripIntegrationRide(t *testing.T, db *sql.DB, riderUserID uuid.UUID) uuid.UUID {
	t.Helper()
	rideID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO ride_requests (
			id, rider_user_id, pickup_latitude, pickup_longitude,
			destination_latitude, destination_longitude, status, proposed_fare_minor, currency
		)
		VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 100000, 'PKR')
	`, rideID, riderUserID); err != nil {
		t.Fatalf("insert ride: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM trips WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_offers WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_requests WHERE id = $1`, rideID)
	})
	return rideID
}

func insertTripIntegrationOffer(t *testing.T, db *sql.DB, rideRequestID, driverUserID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO ride_offers (ride_request_id, driver_user_id, amount_minor, currency)
		VALUES ($1, $2, 100000, 'PKR')
	`, rideRequestID, driverUserID); err != nil {
		t.Fatalf("insert offer: %v", err)
	}
}
