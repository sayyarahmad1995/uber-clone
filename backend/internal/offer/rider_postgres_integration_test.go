package offer

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func TestRiderRejectClosesDriverOpportunity(t *testing.T) {
	db := openOfferIntegrationDB(t)
	ctx := context.Background()
	riderID := createOfferTestUser(t, db, "rider")
	driverID := createOfferTestDriver(t, db)
	rideID := createOfferTestRide(t, db, riderID)
	repository := NewPostgresRepository(db)

	openOfferOpportunity(t, repository, driverID, rideID)
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
	if _, err := repository.Upsert(ctx, rideID, driverID, 115000, 90000, 130000, "PKR"); !errors.Is(err, ErrOpportunityNotOpen) {
		t.Fatalf("same Driver must not re-offer after Rider rejection, got %v", err)
	}
	feed, err := repository.Discover(ctx, driverID, DiscoveryLimit)
	if err != nil {
		t.Fatalf("discover after rejection: %v", err)
	}
	if len(feed) != 0 {
		t.Fatalf("same request must not reappear after Rider rejection, got %#v", feed)
	}
}

func TestDriverOpportunityExpiresOnce(t *testing.T) {
	db := openOfferIntegrationDB(t)
	ctx := context.Background()
	riderID := createOfferTestUser(t, db, "rider")
	driverID := createOfferTestDriver(t, db)
	rideID := createOfferTestRide(t, db, riderID)
	repository := NewPostgresRepository(db)

	openOfferOpportunity(t, repository, driverID, rideID)
	if _, err := db.Exec(`
		UPDATE driver_ride_request_opportunities
		SET opened_at = statement_timestamp() - INTERVAL '60 seconds',
		    visible_until = statement_timestamp() - INTERVAL '30 seconds'
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, driverID); err != nil {
		t.Fatalf("age opportunity: %v", err)
	}
	if _, err := repository.Upsert(ctx, rideID, driverID, 100000, 90000, 130000, "PKR"); !errors.Is(err, ErrOpportunityNotOpen) {
		t.Fatalf("expected closed opportunity after visibility deadline, got %v", err)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM driver_ride_request_opportunities WHERE ride_request_id=$1 AND driver_user_id=$2`, rideID, driverID).Scan(&status); err != nil {
		t.Fatalf("read stale opportunity state: %v", err)
	}
	if status != "open" {
		t.Fatalf("business write materialized opportunity status=%q; periodic sweep must own materialization", status)
	}
	if err := marketplace.NewExpiryService(db).Sweep(ctx); err != nil {
		t.Fatalf("sweep expired opportunity: %v", err)
	}
	if err := db.QueryRow(`SELECT status FROM driver_ride_request_opportunities WHERE ride_request_id=$1 AND driver_user_id=$2`, rideID, driverID).Scan(&status); err != nil {
		t.Fatalf("read opportunity state: %v", err)
	}
	if status != "window_expired" {
		t.Fatalf("expected window_expired, got %q", status)
	}
	feed, err := repository.Discover(ctx, driverID, DiscoveryLimit)
	if err != nil {
		t.Fatalf("discover after window expiry: %v", err)
	}
	if len(feed) != 0 {
		t.Fatalf("expired opportunity must never reopen, got %#v", feed)
	}
}

func TestOfferExpiryIsTerminalForDriverPair(t *testing.T) {
	db := openOfferIntegrationDB(t)
	ctx := context.Background()
	riderID := createOfferTestUser(t, db, "rider")
	driverID := createOfferTestDriver(t, db)
	rideID := createOfferTestRide(t, db, riderID)
	repository := NewPostgresRepository(db)

	openOfferOpportunity(t, repository, driverID, rideID)
	if _, err := repository.Upsert(ctx, rideID, driverID, 100000, 90000, 130000, "PKR"); err != nil {
		t.Fatalf("submit offer: %v", err)
	}
	if _, err := db.Exec(`
		UPDATE ride_offers
		SET created_at = statement_timestamp() - INTERVAL '20 seconds',
		    expires_at = statement_timestamp() - INTERVAL '1 second'
		WHERE ride_request_id=$1 AND driver_user_id=$2
	`, rideID, driverID); err != nil {
		t.Fatalf("age offer: %v", err)
	}
	items, err := repository.ListForRider(ctx, rideID, riderID)
	if err != nil {
		t.Fatalf("list after offer expiry: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expired offer must disappear, got %#v", items)
	}
	if err := marketplace.NewExpiryService(db).Sweep(ctx); err != nil {
		t.Fatalf("sweep expired offer: %v", err)
	}
	offer, err := repository.Get(ctx, rideID, driverID)
	if err != nil {
		t.Fatalf("get expired offer: %v", err)
	}
	if offer.Status != StatusExpired {
		t.Fatalf("expected expired offer, got %q", offer.Status)
	}
	if _, err := repository.Upsert(ctx, rideID, driverID, 105000, 90000, 130000, "PKR"); !errors.Is(err, ErrOpportunityNotOpen) {
		t.Fatalf("expired offer must not permit a second offer, got %v", err)
	}
}

func TestExpiredRideNeverEntersDriverFeed(t *testing.T) {
	db := openOfferIntegrationDB(t)
	ctx := context.Background()
	riderID := createOfferTestUser(t, db, "rider")
	driverID := createOfferTestDriver(t, db)
	rideID := createOfferTestRide(t, db, riderID)
	repository := NewPostgresRepository(db)

	if _, err := db.Exec(`
		UPDATE ride_requests
		SET created_at = statement_timestamp() - INTERVAL '4 minutes',
		    expires_at = statement_timestamp() - INTERVAL '1 second'
		WHERE id=$1
	`, rideID); err != nil {
		t.Fatalf("age ride request: %v", err)
	}
	feed, err := repository.Discover(ctx, driverID, DiscoveryLimit)
	if err != nil {
		t.Fatalf("discover expired ride: %v", err)
	}
	if len(feed) != 0 {
		t.Fatalf("expired ride must not be discoverable, got %#v", feed)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM ride_requests WHERE id=$1`, rideID).Scan(&status); err != nil {
		t.Fatalf("read stale expired ride: %v", err)
	}
	if status != "requested" {
		t.Fatalf("business read materialized ride status=%q; periodic sweep must own materialization", status)
	}
	if err := marketplace.NewExpiryService(db).Sweep(ctx); err != nil {
		t.Fatalf("sweep expired ride: %v", err)
	}
	if err := db.QueryRow(`SELECT status FROM ride_requests WHERE id=$1`, rideID).Scan(&status); err != nil {
		t.Fatalf("read expired ride: %v", err)
	}
	if status != "expired" {
		t.Fatalf("expected ride request expired state, got %q", status)
	}
}

func openOfferOpportunity(t *testing.T, repository PostgresRepository, driverID, rideID uuid.UUID) {
	t.Helper()
	feed, err := repository.Discover(context.Background(), driverID, DiscoveryLimit)
	if err != nil {
		t.Fatalf("discover ride: %v", err)
	}
	if len(feed) != 1 || feed[0].RideRequestID != rideID {
		t.Fatalf("expected one Driver opportunity for ride %s, got %#v", rideID, feed)
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
