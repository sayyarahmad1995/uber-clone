package trip

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPostgresSelectionHasOneWinner(t *testing.T) {
	for _, sameRide := range []bool{true, false} {
		name := "one_driver_two_rides"
		if sameRide {
			name = "two_drivers_one_ride"
		}
		t.Run(name, func(t *testing.T) {
			db := openTripIntegrationDB(t)
			riderID := createTripIntegrationUser(t, db, "rider")
			driverA := createTripIntegrationDriver(t, db)
			driverB := driverA
			if sameRide {
				driverB = createTripIntegrationDriver(t, db)
			}
			rideA := createTripIntegrationRide(t, db, riderID)
			rideB := rideA
			if !sameRide {
				rideB = createTripIntegrationRide(t, db, riderID)
			}
			insertTripIntegrationOffer(t, db, rideA, driverA)
			insertTripIntegrationOffer(t, db, rideB, driverB)
			repository := NewPostgresRepository(db)
			start := make(chan struct{})
			results := make(chan error, 2)
			go func() {
				<-start
				_, err := repository.SelectOffer(context.Background(), rideA, riderID, driverA)
				results <- err
			}()
			go func() {
				<-start
				_, err := repository.SelectOffer(context.Background(), rideB, riderID, driverB)
				results <- err
			}()
			close(start)
			wins := 0
			for i := 0; i < 2; i++ {
				err := <-results
				if err == nil {
					wins++
					continue
				}
				if !errors.Is(err, ErrDriverUnavailable) && !errors.Is(err, ErrMarketplaceNotOpen) {
					t.Fatalf("unexpected selection failure: %v", err)
				}
			}
			if wins != 1 {
				t.Fatalf("expected one winning selection, got %d", wins)
			}
			var count int
			if err := db.QueryRow(`SELECT count(*) FROM trips WHERE ride_request_id IN ($1, $2)`, rideA, rideB).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("expected one persisted Trip, got %d", count)
			}
			if sameRide {
				if err := db.QueryRow(`SELECT count(*) FROM ride_offers WHERE ride_request_id = $1 AND status = 'closed'`, rideA).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 1 {
					t.Fatalf("expected competing offer closed, got %d", count)
				}
			}
		})
	}
}

func TestPostgresSelectionRechecksLocationAfterLockWait(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	insertTripIntegrationOffer(t, db, rideID, driverID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	var blockerPID int
	if err := blocker.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&blockerPID); err != nil {
		t.Fatal(err)
	}
	// Leave the stale update uncommitted while selection starts against the
	// previously fresh location, then make it visible after selection waits.
	if _, err := blocker.ExecContext(ctx, `UPDATE driver_locations SET updated_at=statement_timestamp()-INTERVAL '3 minutes' WHERE driver_user_id=$1`, driverID); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { _, err := NewPostgresRepository(db).SelectOffer(ctx, rideID, riderID, driverID); result <- err }()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, blockerPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("selection did not lock location: %v", err)
		case <-ctx.Done():
			t.Fatal("selection never waited for location lock")
		case <-ticker.C:
		}
	}
	if err := blocker.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, ErrDriverUnavailable) {
			t.Fatalf("selection used pre-wait location: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("selection did not finish after location update")
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM trips WHERE ride_request_id=$1`, rideID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("stale Driver was assigned: %d %v", count, err)
	}
}

func TestPostgresSelectionRevalidatesDriverAvailability(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	insertTripIntegrationOffer(t, db, rideID, driverID)
	if _, err := db.Exec(`UPDATE driver_profiles SET is_online = FALSE WHERE user_id = $1`, driverID); err != nil {
		t.Fatal(err)
	}
	_, err := NewPostgresRepository(db).SelectOffer(context.Background(), rideID, riderID, driverID)
	if !errors.Is(err, ErrDriverUnavailable) {
		t.Fatalf("expected unavailable Driver, got %v", err)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM ride_offers WHERE ride_request_id = $1 AND driver_user_id = $2`, rideID, driverID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("failed selection changed offer: %s", status)
	}
}
