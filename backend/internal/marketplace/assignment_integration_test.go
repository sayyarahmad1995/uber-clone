package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestPostgresAssignmentRejectsInvalidSelection(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query string
		want  error
	}{
		{"stale_version", `UPDATE ride_offers SET updated_at=statement_timestamp() WHERE ride_request_id=$1 RETURNING updated_at`, ErrOfferNotActionable},
		{"expired_ride", `UPDATE ride_requests SET created_at=statement_timestamp()-INTERVAL '10 minutes', expires_at=statement_timestamp()-INTERVAL '1 second' WHERE id=$1`, ErrNotOpen},
		{"non_pending_offer", `UPDATE ride_offers SET status='rejected', decided_at=statement_timestamp() WHERE ride_request_id=$1`, ErrOfferNotActionable},
		{"closed_opportunity", `UPDATE driver_ride_request_opportunities SET status='closed' WHERE ride_request_id=$1`, ErrOfferNotActionable},
		{"context_mismatch", `UPDATE ride_offers SET operation_context=jsonb_set(operation_context,'{service_code}','"comfort"') WHERE ride_request_id=$1`, ErrDriverUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openTripIntegrationDB(t)
			rider := createTripIntegrationUser(t, db, "rider")
			driver := createTripIntegrationDriver(t, db)
			ride := createTripIntegrationRide(t, db, rider)
			version := insertTripIntegrationOffer(t, db, ride, driver)
			if tc.name == "stale_version" {
				var currentVersion time.Time
				if err := db.QueryRow(tc.query, ride).Scan(&currentVersion); err != nil {
					t.Fatal(err)
				}
				if currentVersion.Equal(version) {
					t.Fatal("fixture did not advance persisted offer version")
				}
			} else if _, err := db.Exec(tc.query, ride); err != nil {
				t.Fatal(err)
			}
			_, err := NewPostgresAssignmentRepository(db).SelectOffer(context.Background(), ride, rider, driver, version)
			if !errors.Is(err, tc.want) {
				t.Fatalf("selection error = %v, want %v", err, tc.want)
			}
			var count int
			if err := db.QueryRow(`SELECT count(*) FROM trips WHERE ride_request_id=$1`, ride).Scan(&count); err != nil || count != 0 {
				t.Fatalf("rejected selection created a Trip: count=%d err=%v", count, err)
			}
			var accepted bool
			if err := db.QueryRow(`SELECT status='accepted' FROM ride_requests WHERE id=$1`, ride).Scan(&accepted); err != nil || accepted {
				t.Fatalf("rejected selection accepted ride: accepted=%t err=%v", accepted, err)
			}
		})
	}
}

func TestPostgresAssignmentSuccessClosesCompetingOpportunities(t *testing.T) {
	db := openTripIntegrationDB(t)
	rider := createTripIntegrationUser(t, db, "rider")
	driver := createTripIntegrationDriver(t, db)
	competitor := createTripIntegrationDriver(t, db)
	observer := createTripIntegrationDriver(t, db)
	ride := createTripIntegrationRide(t, db, rider)
	version := insertTripIntegrationOffer(t, db, ride, driver)
	var selectedVehicle uuid.UUID
	if err := db.QueryRow(`SELECT vehicle_id FROM driver_operating_selections WHERE driver_user_id=$1`, driver).Scan(&selectedVehicle); err != nil {
		t.Fatal(err)
	}
	insertTripIntegrationOffer(t, db, ride, competitor)
	if _, err := db.Exec(`INSERT INTO driver_ride_request_opportunities (ride_request_id,driver_user_id,status,visible_until) VALUES ($1,$2,'open',statement_timestamp()+INTERVAL '30 seconds')`, ride, observer); err != nil {
		t.Fatal(err)
	}
	assigned, err := NewPostgresAssignmentRepository(db).SelectOffer(context.Background(), ride, rider, driver, version)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.RideRequestID != ride || assigned.RiderUserID != rider || assigned.DriverUserID != driver || assigned.Status != trip.StatusAssigned {
		t.Fatalf("unexpected assigned Trip: %+v", assigned)
	}
	assertCompleteOperationContext(t, assigned.OperationContext, selectedVehicle)
	var winnerOffer, winnerOpportunity, competingOffer string
	if err := db.QueryRow(`SELECT o.status,p.status FROM ride_offers o JOIN driver_ride_request_opportunities p USING (ride_request_id,driver_user_id) WHERE o.ride_request_id=$1 AND o.driver_user_id=$2`, ride, driver).Scan(&winnerOffer, &winnerOpportunity); err != nil {
		t.Fatal(err)
	}
	if winnerOffer != "accepted" || winnerOpportunity != "accepted" {
		t.Fatalf("winner states: offer=%s opportunity=%s", winnerOffer, winnerOpportunity)
	}
	if err := db.QueryRow(`SELECT status FROM ride_offers WHERE ride_request_id=$1 AND driver_user_id=$2`, ride, competitor).Scan(&competingOffer); err != nil || competingOffer != "closed" {
		t.Fatalf("competing offer=%s err=%v", competingOffer, err)
	}
	for _, other := range []struct {
		name string
		id   any
	}{{"offered", competitor}, {"open", observer}} {
		var status string
		if err := db.QueryRow(`SELECT status FROM driver_ride_request_opportunities WHERE ride_request_id=$1 AND driver_user_id=$2`, ride, other.id).Scan(&status); err != nil || status != "closed" {
			t.Fatalf("competing %s opportunity=%s err=%v", other.name, status, err)
		}
	}
}

func assertCompleteOperationContext(t *testing.T, context *trip.OperationContext, vehicleID uuid.UUID) {
	t.Helper()
	if context == nil {
		t.Fatal("operation context is nil")
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"driver_name":   "Test Driver",
		"vehicle_id":    vehicleID.String(),
		"make":          "Test",
		"model":         "Car",
		"model_year":    float64(2024),
		"color":         "White",
		"license_plate": "XYZ 987",
		"service_code":  "economy",
		"service_name":  "Economy",
		"fare":          map[string]any{"amount_minor": float64(100000), "currency": "PKR"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("assigned operation context mismatch:\n got: %s\nwant: %#v", encoded, want)
	}
}

func TestPostgresAssignmentHasOneWinner(t *testing.T) {
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
			versionA := insertTripIntegrationOffer(t, db, rideA, driverA)
			versionB := insertTripIntegrationOffer(t, db, rideB, driverB)
			repository := NewPostgresAssignmentRepository(db)
			start := make(chan struct{})
			results := make(chan error, 2)
			go func() {
				<-start
				_, err := repository.SelectOffer(
					context.Background(),
					rideA,
					riderID,
					driverA,
					versionA,
				)
				results <- err
			}()
			go func() {
				<-start
				_, err := repository.SelectOffer(
					context.Background(),
					rideB,
					riderID,
					driverB,
					versionB,
				)
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
				if !errors.Is(err, ErrDriverUnavailable) && !errors.Is(err, ErrNotOpen) {
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

func TestPostgresAssignmentRechecksLocationAfterLockWait(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	offerVersion := insertTripIntegrationOffer(t, db, rideID, driverID)
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
	go func() {
		_, err := NewPostgresAssignmentRepository(db).SelectOffer(
			context.Background(),
			rideID,
			riderID,
			driverID,
			offerVersion,
		)
		result <- err
	}()
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

func TestPostgresAssignmentRevalidatesDriverAvailability(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	offerVersion := insertTripIntegrationOffer(t, db, rideID, driverID)
	if _, err := db.Exec(`UPDATE driver_profiles SET is_online = FALSE WHERE user_id = $1`, driverID); err != nil {
		t.Fatal(err)
	}
	_, err := NewPostgresAssignmentRepository(db).SelectOffer(
		context.Background(),
		rideID,
		riderID,
		driverID,
		offerVersion,
	)
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
