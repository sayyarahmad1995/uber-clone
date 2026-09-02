package trip

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
)

const integrationCandidateTimeout = 30 * time.Second

func TestPostgresRepositoryAcceptCreatesAssignedTrip(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createdAt := time.Now().UTC()
	insertTripIntegrationCandidate(t, db, rideID, driverID, createdAt, "pending", nil, nil)

	acceptance, err := NewPostgresRepository(db, integrationCandidateTimeout).Accept(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	if acceptance.Trip.Status != StatusAssigned {
		t.Fatalf("expected assigned trip, got %s", acceptance.Trip.Status)
	}
	if acceptance.Trip.DriverUserID != driverID || acceptance.Trip.RiderUserID != riderID {
		t.Fatalf("unexpected trip participants: %#v", acceptance.Trip)
	}
	if acceptance.CandidateDecidedAt == nil {
		t.Fatal("expected candidate decision timestamp")
	}

	var status string
	var decidedAt sql.NullTime
	var releasedAt sql.NullTime
	if err := db.QueryRow(`
		SELECT status, decided_at, released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, driverID).Scan(&status, &decidedAt, &releasedAt); err != nil {
		t.Fatalf("select candidate: %v", err)
	}
	if status != "accepted" || !decidedAt.Valid || releasedAt.Valid {
		t.Fatalf("unexpected candidate state status=%s decided=%v released=%v", status, decidedAt.Valid, releasedAt.Valid)
	}
}

func TestPostgresRepositoryAcceptExpiredCandidateReleasesWithoutTrip(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createdAt := time.Now().UTC().Add(-45 * time.Second)
	insertTripIntegrationCandidate(t, db, rideID, driverID, createdAt, "pending", nil, nil)

	_, err := NewPostgresRepository(db, integrationCandidateTimeout).Accept(context.Background(), rideID, driverID)
	if !errors.Is(err, ErrAssignmentResolved) {
		t.Fatalf("expected ErrAssignmentResolved, got %v", err)
	}

	var status string
	var releasedAt sql.NullTime
	if err := db.QueryRow(`
		SELECT status, released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, driverID).Scan(&status, &releasedAt); err != nil {
		t.Fatalf("select candidate: %v", err)
	}
	if status != "pending" || !releasedAt.Valid {
		t.Fatalf("expected released pending candidate, got status=%s released=%v", status, releasedAt.Valid)
	}

	var tripCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trips WHERE ride_request_id = $1`, rideID).Scan(&tripCount); err != nil {
		t.Fatalf("count trips: %v", err)
	}
	if tripCount != 0 {
		t.Fatalf("expected no trip, got %d", tripCount)
	}
}

func TestPostgresRepositoryAcceptWrongDriverReturnsNotFoundWithoutMutation(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	matchedDriverID := createTripIntegrationDriver(t, db)
	wrongDriverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createdAt := time.Now().UTC()
	insertTripIntegrationCandidate(t, db, rideID, matchedDriverID, createdAt, "pending", nil, nil)

	_, err := NewPostgresRepository(db, integrationCandidateTimeout).Accept(context.Background(), rideID, wrongDriverID)
	if !errors.Is(err, ErrAssignmentNotFound) {
		t.Fatalf("expected ErrAssignmentNotFound, got %v", err)
	}

	var status string
	var decidedAt sql.NullTime
	var releasedAt sql.NullTime
	if err := db.QueryRow(`
		SELECT status, decided_at, released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, matchedDriverID).Scan(&status, &decidedAt, &releasedAt); err != nil {
		t.Fatalf("select matched candidate: %v", err)
	}
	if status != "pending" || decidedAt.Valid || releasedAt.Valid {
		t.Fatalf("wrong driver mutated matched candidate: status=%s decided=%v released=%v", status, decidedAt.Valid, releasedAt.Valid)
	}
}

func TestPostgresRepositoryAcceptIsIdempotentForAcceptedCandidate(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createdAt := time.Now().UTC()
	insertTripIntegrationCandidate(t, db, rideID, driverID, createdAt, "pending", nil, nil)
	repository := NewPostgresRepository(db, integrationCandidateTimeout)

	first, err := repository.Accept(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("first accept: %v", err)
	}
	second, err := repository.Accept(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("second accept: %v", err)
	}
	if first.CandidateDecidedAt == nil || second.CandidateDecidedAt == nil {
		t.Fatal("expected decision timestamp on both acceptances")
	}
	if !first.CandidateDecidedAt.Equal(*second.CandidateDecidedAt) {
		t.Fatalf("expected decision timestamp to remain stable, first=%v second=%v", *first.CandidateDecidedAt, *second.CandidateDecidedAt)
	}
	if !first.Trip.AssignedAt.Equal(second.Trip.AssignedAt) {
		t.Fatalf("expected assigned_at to remain stable, first=%v second=%v", first.Trip.AssignedAt, second.Trip.AssignedAt)
	}

	var tripCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM trips WHERE ride_request_id = $1`, rideID).Scan(&tripCount); err != nil {
		t.Fatalf("count trips: %v", err)
	}
	if tripCount != 1 {
		t.Fatalf("expected exactly one trip, got %d", tripCount)
	}
}

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
	return userID
}

func createTripIntegrationRide(t *testing.T, db *sql.DB, riderUserID uuid.UUID) uuid.UUID {
	t.Helper()
	rideID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO ride_requests (
			id, rider_user_id, pickup_latitude, pickup_longitude,
			destination_latitude, destination_longitude, status, booking_mode
		)
		VALUES ($1, $2, 24.8610, 67.0010, 24.8800, 67.0200, 'requested', 'automatic')
	`, rideID, riderUserID); err != nil {
		t.Fatalf("insert ride: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM trips WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_driver_candidates WHERE ride_request_id = $1`, rideID)
		_, _ = db.Exec(`DELETE FROM ride_requests WHERE id = $1`, rideID)
	})
	return rideID
}

func insertTripIntegrationCandidate(t *testing.T, db *sql.DB, rideRequestID, driverUserID uuid.UUID, createdAt time.Time, status string, decidedAt, releasedAt *time.Time) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO ride_driver_candidates (
			ride_request_id, driver_user_id, status, created_at, decided_at, released_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, rideRequestID, driverUserID, status, createdAt, decidedAt, releasedAt); err != nil {
		t.Fatalf("insert candidate: %v", err)
	}
}
