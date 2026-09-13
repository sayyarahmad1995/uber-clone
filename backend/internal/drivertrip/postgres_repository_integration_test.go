package drivertrip

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func openDriverTripIntegrationDB(t *testing.T) *sql.DB {
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

func createDriverTripUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
	return userID
}

func createDriverTripRide(t *testing.T, db *sql.DB, riderUserID uuid.UUID) uuid.UUID {
	t.Helper()
	rideID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO ride_requests (
			id, rider_user_id, pickup_latitude, pickup_longitude,
			destination_latitude, destination_longitude, status, proposed_fare_minor, currency
		)
		VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 100000, 'PKR')
	`, rideID, riderUserID); err != nil {
		t.Fatalf("insert ride request: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM trips WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_requests WHERE id = $1`, rideID)
	})
	return rideID
}

func createCompletedDriverTrip(t *testing.T, db *sql.DB, rideID, riderID, driverID uuid.UUID) {
	t.Helper()
	assignedAt := time.Now().UTC().Add(-10 * time.Minute)
	completedAt := time.Now().UTC().Add(-1 * time.Minute)
	if _, err := db.Exec(`
		INSERT INTO trips (
			ride_request_id, rider_user_id, driver_user_id,
			status, assigned_at, started_at, completed_at, operation_context
		)
		VALUES ($1, $2, $3, 'completed', $4, $4, $5, '{"fare":{"amount_minor":100000,"currency":"PKR"}}')
	`, rideID, riderID, driverID, assignedAt, completedAt); err != nil {
		t.Fatalf("insert completed trip: %v", err)
	}
}

func createCashCollectedDriverTrip(t *testing.T, db *sql.DB, rideID, riderID, driverID uuid.UUID) {
	t.Helper()
	assignedAt := time.Now().UTC().Add(-10 * time.Minute)
	completedAt := time.Now().UTC().Add(-1 * time.Minute)
	if _, err := db.Exec(`
		INSERT INTO trips (
			ride_request_id, rider_user_id, driver_user_id,
			status, assigned_at, started_at, completed_at, operation_context,
			settlement_status, settlement_method, cash_collected_at, cash_collected_by
		)
		VALUES (
			$1, $2, $3, 'completed', $4, $4, $5,
			'{"fare":{"amount_minor":100000,"currency":"PKR"}}',
			'cash_collected', 'cash', $5, $3
		)
	`, rideID, riderID, driverID, assignedAt, completedAt); err != nil {
		t.Fatalf("insert cash-collected trip: %v", err)
	}
}

func TestPostgresRepositoryGetCurrentKeepsCompletedUnsettledTrip(t *testing.T) {
	db := openDriverTripIntegrationDB(t)
	riderID := createDriverTripUser(t, db)
	driverID := createDriverTripUser(t, db)
	rideID := createDriverTripRide(t, db, riderID)
	createCompletedDriverTrip(t, db, rideID, riderID, driverID)

	view, err := NewPostgresRepository(db).GetCurrent(context.Background(), driverID)
	if err != nil {
		t.Fatalf("GetCurrent() error = %v", err)
	}
	if view.RideRequestID != rideID {
		t.Fatalf("ride request id = %s, want %s", view.RideRequestID, rideID)
	}
	if view.Status != trip.StatusCompleted {
		t.Fatalf("status = %s, want %s", view.Status, trip.StatusCompleted)
	}
	if view.Settlement.Status != trip.SettlementUnsettled {
		t.Fatalf("settlement = %s, want %s", view.Settlement.Status, trip.SettlementUnsettled)
	}
}

func TestPostgresRepositoryGetCurrentExcludesCashCollectedTrip(t *testing.T) {
	db := openDriverTripIntegrationDB(t)
	riderID := createDriverTripUser(t, db)
	driverID := createDriverTripUser(t, db)
	rideID := createDriverTripRide(t, db, riderID)
	createCashCollectedDriverTrip(t, db, rideID, riderID, driverID)

	_, err := NewPostgresRepository(db).GetCurrent(context.Background(), driverID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetCurrent() error = %v, want %v", err, ErrNotFound)
	}
}
