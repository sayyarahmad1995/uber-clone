package offer

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func TestListForRiderHidesRejectedOfferUntilDriverResubmits(t *testing.T) {
	db := openOfferIntegrationDB(t)
	ctx := context.Background()
	riderID := createOfferTestUser(t, db, "rider")
	driverID := createOfferTestDriver(t, db)
	rideID := createOfferTestRide(t, db, riderID)
	repository := NewPostgresRepository(db)

	if _, err := repository.Upsert(ctx, rideID, driverID, 110000, 90000, 130000, "PKR"); err != nil {
		t.Fatalf("submit offer: %v", err)
	}
	items, err := repository.ListForRider(ctx, rideID, riderID)
	if err != nil {
		t.Fatalf("list pending offer: %v", err)
	}
	if len(items) != 1 || items[0].Status != StatusPending || items[0].AmountMinor != 110000 {
		t.Fatalf("expected pending offer in Rider comparison, got %#v", items)
	}

	rejected, err := repository.Reject(ctx, rideID, riderID, driverID)
	if err != nil {
		t.Fatalf("reject offer: %v", err)
	}
	if rejected.Status != StatusRejected {
		t.Fatalf("expected rejected status, got %s", rejected.Status)
	}
	items, err = repository.ListForRider(ctx, rideID, riderID)
	if err != nil {
		t.Fatalf("list after rejection: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("rejected offer must disappear from active Rider comparison, got %#v", items)
	}

	if _, err := repository.Upsert(ctx, rideID, driverID, 115000, 90000, 130000, "PKR"); err != nil {
		t.Fatalf("resubmit offer: %v", err)
	}
	items, err = repository.ListForRider(ctx, rideID, riderID)
	if err != nil {
		t.Fatalf("list resubmitted offer: %v", err)
	}
	if len(items) != 1 || items[0].Status != StatusPending || items[0].AmountMinor != 115000 {
		t.Fatalf("expected resubmitted pending offer to reappear, got %#v", items)
	}
}

func openOfferIntegrationDB(t *testing.T) *sql.DB {
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

func createOfferTestUser(t *testing.T, db *sql.DB, capability string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_capabilities (user_id, capability) VALUES ($1, $2)`, userID, capability); err != nil {
		t.Fatalf("insert capability: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID) })
	return userID
}

func createOfferTestDriver(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := createOfferTestUser(t, db, "driver")
	if _, err := db.Exec(`INSERT INTO driver_profiles (user_id, status, is_online, display_name) VALUES ($1, 'active', TRUE, 'Test Driver')`, userID); err != nil {
		t.Fatalf("insert driver profile: %v", err)
	}
	vehicleID := uuid.New()
	if _, err := db.Exec(`INSERT INTO driver_vehicles (id, driver_user_id, make, model, model_year, color, license_plate) VALUES ($1, $2, 'Test', 'Car', 2024, 'White', 'TEST')`, vehicleID, userID); err != nil {
		t.Fatalf("insert vehicle: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO driver_vehicle_service_enrollments (vehicle_id, service_code, approved_at, approved_by) VALUES ($1, 'economy', NOW(), 'test-reviewer')`, vehicleID); err != nil {
		t.Fatalf("insert service enrollment: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO driver_operating_selections (driver_user_id, vehicle_id, service_code) VALUES ($1, $2, 'economy')`, userID, vehicleID); err != nil {
		t.Fatalf("insert operating selection: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO driver_locations (driver_user_id, latitude, longitude, updated_at) VALUES ($1, 24.86, 67.0, NOW())`, userID); err != nil {
		t.Fatalf("insert location: %v", err)
	}
	return userID
}

func createOfferTestRide(t *testing.T, db *sql.DB, riderID uuid.UUID) uuid.UUID {
	t.Helper()
	rideID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO ride_requests (id, rider_user_id, pickup_latitude, pickup_longitude,
		    destination_latitude, destination_longitude, status, proposed_fare_minor, currency, service_code)
		VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 100000, 'PKR', 'economy')
	`, rideID, riderID); err != nil {
		t.Fatalf("insert ride: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM ride_offers WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_requests WHERE id = $1`, rideID)
	})
	return rideID
}
