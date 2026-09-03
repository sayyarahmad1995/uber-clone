package trip

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPostgresRepositoryStartTransitionsAssignedTrip(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	started, err := repository.Start(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if started.Status != StatusInProgress {
		t.Fatalf("expected in_progress, got %s", started.Status)
	}
	if started.StartedAt == nil {
		t.Fatal("expected started_at")
	}
	if started.CompletedAt != nil || started.CancelledAt != nil {
		t.Fatalf("unexpected terminal timestamps: completed=%v cancelled=%v", started.CompletedAt, started.CancelledAt)
	}
}

func TestPostgresRepositoryStartIsIdempotent(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	first, err := repository.Start(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("first start: %v", err)
	}
	second, err := repository.Start(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("second start: %v", err)
	}
	if first.StartedAt == nil || second.StartedAt == nil {
		t.Fatal("expected started_at on both results")
	}
	if !first.StartedAt.Equal(*second.StartedAt) {
		t.Fatalf("expected stable started_at, first=%v second=%v", *first.StartedAt, *second.StartedAt)
	}
}

func TestPostgresRepositoryCompleteTransitionsAndReleasesCandidate(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	if _, err := repository.Start(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("start: %v", err)
	}
	completed, err := repository.Complete(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if completed.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", completed.Status)
	}
	if completed.CompletedAt == nil {
		t.Fatal("expected completed_at")
	}

	var releasedAt sql.NullTime
	if err := db.QueryRow(`
		SELECT released_at
		FROM ride_driver_candidates
		WHERE ride_request_id = $1 AND driver_user_id = $2
	`, rideID, driverID).Scan(&releasedAt); err != nil {
		t.Fatalf("select candidate release: %v", err)
	}
	if !releasedAt.Valid {
		t.Fatal("expected accepted candidate to be released on completion")
	}
	if !releasedAt.Time.Equal(*completed.CompletedAt) {
		t.Fatalf("expected released_at to equal completed_at, released=%v completed=%v", releasedAt.Time, *completed.CompletedAt)
	}
}

func TestPostgresRepositoryCompleteBeforeStartIsRejected(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	_, err := NewPostgresRepository(db).Complete(context.Background(), rideID, driverID)
	if !errors.Is(err, ErrTripNotStarted) {
		t.Fatalf("expected ErrTripNotStarted, got %v", err)
	}

	var status string
	var completedAt sql.NullTime
	if err := db.QueryRow(`SELECT status, completed_at FROM trips WHERE ride_request_id = $1`, rideID).Scan(&status, &completedAt); err != nil {
		t.Fatalf("select trip: %v", err)
	}
	if status != string(StatusAssigned) || completedAt.Valid {
		t.Fatalf("complete-before-start mutated trip: status=%s completed=%v", status, completedAt.Valid)
	}
}

func TestPostgresRepositoryStartAfterCompletionIsRejected(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	if _, err := repository.Start(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("start: %v", err)
	}
	completed, err := repository.Complete(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	_, err = repository.Start(context.Background(), rideID, driverID)
	if !errors.Is(err, ErrTripCompleted) {
		t.Fatalf("expected ErrTripCompleted, got %v", err)
	}

	var completedAt sql.NullTime
	if err := db.QueryRow(`SELECT completed_at FROM trips WHERE ride_request_id = $1`, rideID).Scan(&completedAt); err != nil {
		t.Fatalf("select trip completion: %v", err)
	}
	if !completedAt.Valid || !completedAt.Time.Equal(*completed.CompletedAt) {
		t.Fatalf("terminal state changed unexpectedly: db=%v result=%v", completedAt, completed.CompletedAt)
	}
}

func TestPostgresRepositoryCompleteIsIdempotent(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	if _, err := repository.Start(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("start: %v", err)
	}
	first, err := repository.Complete(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("first complete: %v", err)
	}
	second, err := repository.Complete(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("second complete: %v", err)
	}
	if first.CompletedAt == nil || second.CompletedAt == nil {
		t.Fatal("expected completed_at on both results")
	}
	if !first.CompletedAt.Equal(*second.CompletedAt) {
		t.Fatalf("expected stable completed_at, first=%v second=%v", *first.CompletedAt, *second.CompletedAt)
	}
}

func createAcceptedTripFixture(t *testing.T, db *sql.DB, rideID, riderID, driverID uuid.UUID) {
	t.Helper()
	createdAt := time.Now().UTC().Add(-time.Second)
	decidedAt := time.Now().UTC()
	insertTripIntegrationCandidate(t, db, rideID, driverID, createdAt, "accepted", &decidedAt, nil)
	if _, err := db.Exec(`
		INSERT INTO trips (ride_request_id, rider_user_id, driver_user_id, status, assigned_at)
		VALUES ($1, $2, $3, 'assigned', $4)
	`, rideID, riderID, driverID, decidedAt); err != nil {
		t.Fatalf("insert assigned trip: %v", err)
	}
}
