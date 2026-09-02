package matching

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func TestPostgresRepositoryMatchSelectsNearestFreshDriver(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	nearDriverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC()))
	farDriverID := createIntegrationDriver(t, db, locationAt(24.9000, 67.0500, time.Now().UTC()))
	rideID := createIntegrationRide(t, db, riderID, "automatic")

	result, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Match(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if !result.Created {
		t.Fatal("expected a newly created candidate")
	}
	if result.Candidate.DriverUserID != nearDriverID {
		t.Fatalf("expected nearest driver %s, got %s (far driver %s)", nearDriverID, result.Candidate.DriverUserID, farDriverID)
	}
}

func TestPostgresRepositoryMatchExcludesStaleAndMissingLocations(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	staleDriverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC().Add(-5*time.Minute)))
	missingLocationDriverID := createIntegrationDriver(t, db, nil)
	freshDriverID := createIntegrationDriver(t, db, locationAt(24.8800, 67.0200, time.Now().UTC()))
	rideID := createIntegrationRide(t, db, riderID, "automatic")

	result, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Match(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if result.Candidate.DriverUserID != freshDriverID {
		t.Fatalf("expected fresh driver %s, got %s (stale=%s missing=%s)", freshDriverID, result.Candidate.DriverUserID, staleDriverID, missingLocationDriverID)
	}
}

func TestPostgresRepositoryMatchBlocksUnreleasedCandidateButNotReleasedHistory(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	blockedDriverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC()))
	availableDriverID := createIntegrationDriver(t, db, locationAt(24.8800, 67.0200, time.Now().UTC()))
	blockingRideID := createIntegrationRide(t, db, riderID, "automatic")
	insertCandidate(t, db, blockingRideID, blockedDriverID, time.Now().UTC(), nil)
	firstRideID := createIntegrationRide(t, db, riderID, "automatic")

	first, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Match(context.Background(), firstRideID, riderID)
	if err != nil {
		t.Fatalf("match with blocked driver: %v", err)
	}
	if first.Candidate.DriverUserID != availableDriverID {
		t.Fatalf("expected available driver %s, got %s", availableDriverID, first.Candidate.DriverUserID)
	}

	releasedAt := time.Now().UTC()
	if _, err := db.Exec(`UPDATE ride_driver_candidates SET released_at = $3 WHERE ride_request_id = $1 AND driver_user_id = $2`, blockingRideID, blockedDriverID, releasedAt); err != nil {
		t.Fatalf("release blocking candidate: %v", err)
	}
	if _, err := db.Exec(`UPDATE ride_driver_candidates SET released_at = $3 WHERE ride_request_id = $1 AND driver_user_id = $2`, firstRideID, availableDriverID, releasedAt); err != nil {
		t.Fatalf("release first match candidate: %v", err)
	}

	secondRideID := createIntegrationRide(t, db, riderID, "automatic")
	second, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Match(context.Background(), secondRideID, riderID)
	if err != nil {
		t.Fatalf("match after release: %v", err)
	}
	if second.Candidate.DriverUserID != blockedDriverID {
		t.Fatalf("expected released-history driver %s to become eligible, got %s", blockedDriverID, second.Candidate.DriverUserID)
	}
}

func TestPostgresRepositoryMatchReleasesExpiredPendingCandidate(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	driverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC()))
	oldRideID := createIntegrationRide(t, db, riderID, "automatic")
	oldCreatedAt := time.Now().UTC().Add(-45 * time.Second)
	insertCandidate(t, db, oldRideID, driverID, oldCreatedAt, nil)
	newRideID := createIntegrationRide(t, db, riderID, "automatic")

	result, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Match(context.Background(), newRideID, riderID)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if result.Candidate.DriverUserID != driverID {
		t.Fatalf("expected expired-candidate driver %s, got %s", driverID, result.Candidate.DriverUserID)
	}

	var releasedAt sql.NullTime
	if err := db.QueryRow(`SELECT released_at FROM ride_driver_candidates WHERE ride_request_id = $1 AND driver_user_id = $2`, oldRideID, driverID).Scan(&releasedAt); err != nil {
		t.Fatalf("select old candidate release: %v", err)
	}
	if !releasedAt.Valid {
		t.Fatal("expected expired pending candidate to be released")
	}
}

func TestPostgresRepositoryMatchReturnsExistingCandidateIdempotently(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	driverID := createIntegrationDriver(t, db, locationAt(24.8611, 67.0011, time.Now().UTC()))
	rideID := createIntegrationRide(t, db, riderID, "automatic")
	repository := NewPostgresRepository(db, 2*time.Minute, 30*time.Second)

	first, err := repository.Match(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("first match: %v", err)
	}
	if first.Candidate.DriverUserID != driverID || !first.Created {
		t.Fatalf("unexpected first result: %#v", first)
	}

	second, err := repository.Match(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("second match: %v", err)
	}
	if second.Created {
		t.Fatal("expected existing candidate on second match")
	}
	if second.Candidate.DriverUserID != first.Candidate.DriverUserID || !second.Candidate.CreatedAt.Equal(first.Candidate.CreatedAt) {
		t.Fatalf("expected same candidate, first=%#v second=%#v", first.Candidate, second.Candidate)
	}
}

func TestPostgresRepositoryMatchRejectsOffersMode(t *testing.T) {
	db := openIntegrationDB(t)
	riderID := createIntegrationRider(t, db)
	rideID := createIntegrationRide(t, db, riderID, "offers")

	_, err := NewPostgresRepository(db, 2*time.Minute, 30*time.Second).Match(context.Background(), rideID, riderID)
	if !errors.Is(err, ErrRideNotMatchable) {
		t.Fatalf("expected ErrRideNotMatchable, got %v", err)
	}
}

type integrationLocation struct {
	latitude  float64
	longitude float64
	updatedAt time.Time
}

func locationAt(latitude, longitude float64, updatedAt time.Time) *integrationLocation {
	return &integrationLocation{latitude: latitude, longitude: longitude, updatedAt: updatedAt}
}

func openIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
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

func createIntegrationRider(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	insertIntegrationUser(t, db, userID, "rider")
	return userID
}

func createIntegrationDriver(t *testing.T, db *sql.DB, location *integrationLocation) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	insertIntegrationUser(t, db, userID, "driver")
	if _, err := db.Exec(`INSERT INTO driver_profiles (user_id, status, is_online) VALUES ($1, 'active', TRUE)`, userID); err != nil {
		t.Fatalf("insert driver profile: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO driver_vehicles (id, driver_user_id, make, model, color, license_plate)
		VALUES ($1, $2, 'Toyota', 'Corolla', 'White', $3)
	`, uuid.New(), userID, fmt.Sprintf("T-%s", userID.String()[:8])); err != nil {
		t.Fatalf("insert driver vehicle: %v", err)
	}
	if location != nil {
		if _, err := db.Exec(`
			INSERT INTO driver_locations (driver_user_id, latitude, longitude, updated_at)
			VALUES ($1, $2, $3, $4)
		`, userID, location.latitude, location.longitude, location.updatedAt); err != nil {
			t.Fatalf("insert driver location: %v", err)
		}
	}
	return userID
}

func insertIntegrationUser(t *testing.T, db *sql.DB, userID uuid.UUID, capability string) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO users (id, external_subject) VALUES ($1, $2)`, userID, "integration:"+userID.String()); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_capabilities (user_id, capability) VALUES ($1, $2)`, userID, capability); err != nil {
		t.Fatalf("insert capability: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
}

func createIntegrationRide(t *testing.T, db *sql.DB, riderUserID uuid.UUID, bookingMode string) uuid.UUID {
	t.Helper()
	rideID := uuid.New()
	if bookingMode == "offers" {
		if _, err := db.Exec(`
			INSERT INTO ride_requests (
				id, rider_user_id, pickup_latitude, pickup_longitude,
				destination_latitude, destination_longitude, status,
				booking_mode, proposed_fare_minor, currency
			)
			VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 'offers', 100000, 'PKR')
		`, rideID, riderUserID); err != nil {
			t.Fatalf("insert offers ride: %v", err)
		}
	} else {
		if _, err := db.Exec(`
			INSERT INTO ride_requests (
				id, rider_user_id, pickup_latitude, pickup_longitude,
				destination_latitude, destination_longitude, status, booking_mode
			)
			VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 'automatic')
		`, rideID, riderUserID); err != nil {
			t.Fatalf("insert automatic ride: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM ride_requests WHERE id = $1`, rideID)
	})
	return rideID
}

func insertCandidate(t *testing.T, db *sql.DB, rideRequestID, driverUserID uuid.UUID, createdAt time.Time, releasedAt *time.Time) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO ride_driver_candidates (ride_request_id, driver_user_id, status, created_at, released_at)
		VALUES ($1, $2, 'pending', $3, $4)
	`, rideRequestID, driverUserID, createdAt, releasedAt); err != nil {
		t.Fatalf("insert candidate: %v", err)
	}
}
