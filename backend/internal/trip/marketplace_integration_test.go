package trip

import (
	"context"
	"errors"
	"testing"
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
