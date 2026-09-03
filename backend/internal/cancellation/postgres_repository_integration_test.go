package cancellation

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
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestPostgresRepositoryCancelByRiderClosesPendingOffer(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	insertCancellationOffer(t, db, rideID, driverID)

	result, err := NewPostgresRepository(db).CancelByRider(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("cancel by rider: %v", err)
	}
	if result.Status != ride.StatusCancelled || result.CancelledBy != ride.CancellationActorRider || result.Trip != nil {
		t.Fatalf("unexpected cancellation result: %#v", result)
	}
	assertRideCancelledAt(t, db, rideID, "rider", result.CancelledAt)
	assertDriverAvailable(t, db, riderID, driverID)
}

func TestPostgresRepositoryCancelByRiderCancelsAssignedTripAndFreesDriver(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	insertCancellationTrip(t, db, rideID, riderID, driverID, trip.StatusAssigned)

	result, err := NewPostgresRepository(db).CancelByRider(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("cancel by rider: %v", err)
	}
	if result.Trip == nil || result.Trip.Status != trip.StatusCancelled || result.Trip.CancelledAt == nil {
		t.Fatalf("expected cancelled trip, got %#v", result.Trip)
	}
	if !result.Trip.CancelledAt.Equal(result.CancelledAt) {
		t.Fatalf("expected ride/trip cancellation timestamps to match, ride=%v trip=%v", result.CancelledAt, result.Trip.CancelledAt)
	}
	assertDriverAvailable(t, db, riderID, driverID)
}

func TestPostgresRepositoryCancelByDriverCancelsInProgressTrip(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	insertCancellationTrip(t, db, rideID, riderID, driverID, trip.StatusInProgress)

	result, err := NewPostgresRepository(db).CancelByDriver(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("cancel by driver: %v", err)
	}
	if result.CancelledBy != ride.CancellationActorDriver || result.Trip == nil || result.Trip.Status != trip.StatusCancelled {
		t.Fatalf("unexpected driver cancellation result: %#v", result)
	}
	assertRideCancelledAt(t, db, rideID, "driver", result.CancelledAt)
	assertDriverAvailable(t, db, riderID, driverID)
}

func TestPostgresRepositoryCancelIsIdempotent(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	insertCancellationTrip(t, db, rideID, riderID, driverID, trip.StatusAssigned)
	repository := NewPostgresRepository(db)

	first, err := repository.CancelByRider(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("first cancel: %v", err)
	}
	second, err := repository.CancelByRider(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("second cancel: %v", err)
	}
	if first.CancelledBy != second.CancelledBy || !first.CancelledAt.Equal(second.CancelledAt) {
		t.Fatalf("expected stable cancellation metadata, first=%#v second=%#v", first, second)
	}
	if first.Trip == nil || second.Trip == nil || first.Trip.CancelledAt == nil || second.Trip.CancelledAt == nil {
		t.Fatal("expected cancelled trip on both results")
	}
	if !first.Trip.CancelledAt.Equal(*second.Trip.CancelledAt) {
		t.Fatalf("expected stable trip cancelled_at, first=%v second=%v", first.Trip.CancelledAt, second.Trip.CancelledAt)
	}
}

func TestPostgresRepositoryCancelCompletedTripIsRejectedWithoutMutation(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	insertCancellationTrip(t, db, rideID, riderID, driverID, trip.StatusCompleted)

	var beforeCompletedAt time.Time
	if err := db.QueryRow(`SELECT completed_at FROM trips WHERE ride_request_id = $1`, rideID).Scan(&beforeCompletedAt); err != nil {
		t.Fatalf("select trip completion before cancellation: %v", err)
	}

	_, err := NewPostgresRepository(db).CancelByRider(context.Background(), rideID, riderID)
	if !errors.Is(err, ErrTripCompleted) {
		t.Fatalf("expected ErrTripCompleted, got %v", err)
	}

	var status string
	var cancelledAt sql.NullTime
	if err := db.QueryRow(`SELECT status, cancelled_at FROM ride_requests WHERE id = $1`, rideID).Scan(&status, &cancelledAt); err != nil {
		t.Fatalf("select ride: %v", err)
	}
	if status != string(ride.StatusRequested) || cancelledAt.Valid {
		t.Fatalf("completed-trip cancellation mutated ride: status=%s cancelled=%v", status, cancelledAt.Valid)
	}
	var persistedCompletedAt time.Time
	if err := db.QueryRow(`SELECT completed_at FROM trips WHERE ride_request_id = $1`, rideID).Scan(&persistedCompletedAt); err != nil {
		t.Fatalf("select trip completion after cancellation: %v", err)
	}
	if !persistedCompletedAt.Equal(beforeCompletedAt) {
		t.Fatalf("completed_at changed unexpectedly: before=%v after=%v", beforeCompletedAt, persistedCompletedAt)
	}
}

func TestPostgresRepositoryCancelClosesPendingOffersWithoutTouchingOtherRide(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	otherDriverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	otherRideID := createCancellationRide(t, db, riderID)
	insertCancellationOffer(t, db, rideID, driverID)
	insertCancellationOffer(t, db, otherRideID, otherDriverID)

	result, err := NewPostgresRepository(db).CancelByRider(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("cancel offers ride: %v", err)
	}

	var status string
	var decidedAt sql.NullTime
	if err := db.QueryRow(`SELECT status, decided_at FROM ride_offers WHERE ride_request_id = $1 AND driver_user_id = $2`, rideID, driverID).Scan(&status, &decidedAt); err != nil {
		t.Fatalf("select cancelled ride offer: %v", err)
	}
	if status != "closed" || !decidedAt.Valid || !decidedAt.Time.Equal(result.CancelledAt) {
		t.Fatalf("expected closed offer at cancellation time, status=%s decided=%v cancelled=%v", status, decidedAt, result.CancelledAt)
	}

	var otherStatus string
	var otherDecidedAt sql.NullTime
	if err := db.QueryRow(`SELECT status, decided_at FROM ride_offers WHERE ride_request_id = $1 AND driver_user_id = $2`, otherRideID, otherDriverID).Scan(&otherStatus, &otherDecidedAt); err != nil {
		t.Fatalf("select unrelated offer: %v", err)
	}
	if otherStatus != "pending" || otherDecidedAt.Valid {
		t.Fatalf("unrelated offer was mutated: status=%s decided=%v", otherStatus, otherDecidedAt)
	}
}

func openCancellationIntegrationDB(t *testing.T) *sql.DB {
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

func createCancellationUser(t *testing.T, db *sql.DB, capability string) uuid.UUID {
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

func createCancellationDriver(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := createCancellationUser(t, db, "driver")
	if _, err := db.Exec(`INSERT INTO driver_profiles (user_id, status, is_online) VALUES ($1, 'active', TRUE)`, userID); err != nil {
		t.Fatalf("insert driver profile: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO driver_vehicles (id, driver_user_id, make, model, color, license_plate) VALUES ($1, $2, 'Test', 'Car', 'White', 'TEST')`, uuid.New(), userID); err != nil {
		t.Fatalf("insert vehicle: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO driver_locations (driver_user_id, latitude, longitude, updated_at) VALUES ($1, 24.86, 67.0, NOW())`, userID); err != nil {
		t.Fatalf("insert location: %v", err)
	}
	return userID
}

func createCancellationRide(t *testing.T, db *sql.DB, riderID uuid.UUID) uuid.UUID {
	t.Helper()
	rideID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO ride_requests (id, rider_user_id, pickup_latitude, pickup_longitude,
		    destination_latitude, destination_longitude, status, proposed_fare_minor, currency)
		VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 100000, 'PKR')
	`, rideID, riderID); err != nil {
		t.Fatalf("insert ride: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM trips WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_offers WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_requests WHERE id = $1`, rideID)
	})
	return rideID
}

func insertCancellationTrip(t *testing.T, db *sql.DB, rideID, riderID, driverID uuid.UUID, status trip.Status) time.Time {
	t.Helper()
	insertCancellationOffer(t, db, rideID, driverID)
	repository := trip.NewPostgresRepository(db)
	if _, err := repository.SelectOffer(context.Background(), rideID, riderID, driverID); err != nil {
		t.Fatalf("select offer: %v", err)
	}
	if status == trip.StatusInProgress || status == trip.StatusCompleted {
		if _, err := repository.Start(context.Background(), rideID, driverID); err != nil {
			t.Fatalf("start: %v", err)
		}
	}
	if status == trip.StatusCompleted {
		completed, err := repository.Complete(context.Background(), rideID, driverID)
		if err != nil {
			t.Fatalf("complete: %v", err)
		}
		return *completed.CompletedAt
	}
	return time.Time{}
}
func insertCancellationOffer(t *testing.T, db *sql.DB, rideID, driverID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO ride_offers (ride_request_id, driver_user_id, amount_minor, currency, status)
		VALUES ($1, $2, 90000, 'PKR', 'pending')
	`, rideID, driverID); err != nil {
		t.Fatalf("insert ride offer: %v", err)
	}
}

func assertRideCancelledAt(t *testing.T, db *sql.DB, rideID uuid.UUID, actor string, cancelledAt time.Time) {
	t.Helper()
	var status, cancelledBy string
	var persistedCancelledAt time.Time
	if err := db.QueryRow(`SELECT status, cancelled_by, cancelled_at FROM ride_requests WHERE id = $1`, rideID).Scan(&status, &cancelledBy, &persistedCancelledAt); err != nil {
		t.Fatalf("select cancelled ride: %v", err)
	}
	if status != string(ride.StatusCancelled) || cancelledBy != actor || !persistedCancelledAt.Equal(cancelledAt) {
		t.Fatalf("unexpected cancelled ride state status=%s actor=%s at=%v", status, cancelledBy, persistedCancelledAt)
	}
}

func assertDriverAvailable(t *testing.T, db *sql.DB, riderID, driverID uuid.UUID) {
	t.Helper()
	nextRideID := createCancellationRide(t, db, riderID)
	insertCancellationTrip(t, db, nextRideID, riderID, driverID, trip.StatusAssigned)
}
