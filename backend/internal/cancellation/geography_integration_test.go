package cancellation

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func geoExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}

func geoOffers(db *sql.DB) offer.Service {
	return offer.NewService(offer.NewPostgresRepository(db), trip.NewService(trip.NewPostgresRepository(db)))
}

func TestGeographicDiscoveryRanksBeforeLimiting(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	driver := createCancellationDriver(t, db)
	geoExec(t, db, `UPDATE driver_locations SET latitude=0, longitude=0 WHERE driver_user_id=$1`, driver)
	nearest := createCancellationRide(t, db, rider)
	geoExec(t, db, `UPDATE ride_requests SET pickup_latitude=0, pickup_longitude=0, created_at=NOW()-INTERVAL '1 day' WHERE id=$1`, nearest)
	for i := 0; i < offer.DiscoveryLimit+1; i++ {
		id := createCancellationRide(t, db, rider)
		geoExec(t, db, `UPDATE ride_requests SET pickup_latitude=0, pickup_longitude=1 WHERE id=$1`, id)
	}
	self := createCancellationRide(t, db, driver)
	geoExec(t, db, `UPDATE ride_requests SET pickup_latitude=0, pickup_longitude=0 WHERE id=$1`, self)
	items, err := geoOffers(db).Discover(ctx, driver)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != offer.DiscoveryLimit || items[0].RideRequestID != nearest || items[0].PickupDistanceMeters != 0 {
		t.Fatalf("nearest old request missing from limited feed: %+v", items)
	}
	for _, item := range items {
		if item.RideRequestID == self {
			t.Fatal("Driver discovered own ride")
		}
		if item.RideRequestID != nearest && math.Abs(item.PickupDistanceMeters-111194.9266) > 1 {
			t.Fatalf("unexpected distance: %f", item.PickupDistanceMeters)
		}
	}
}

func TestGeographicDiscoveryTieBreaksAndGlobalDistances(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	rider := createCancellationUser(t, db, "rider")
	driver := createCancellationDriver(t, db)
	a, b := createCancellationRide(t, db, rider), createCancellationRide(t, db, rider)
	geoExec(t, db, `UPDATE driver_locations SET latitude=0, longitude=179.9 WHERE driver_user_id=$1`, driver)
	geoExec(t, db, `UPDATE ride_requests SET pickup_latitude=0, pickup_longitude=-179.9, created_at='2026-01-01' WHERE id IN ($1,$2)`, a, b)
	items, err := geoOffers(db).Discover(context.Background(), driver)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].RideRequestID.String() > items[1].RideRequestID.String() {
		t.Fatalf("non-deterministic UUID ties: %+v", items)
	}
	if math.Abs(items[0].PickupDistanceMeters-22238.9853) > 1 {
		t.Fatalf("antimeridian distance: %f", items[0].PickupDistanceMeters)
	}
	geoExec(t, db, `UPDATE ride_requests SET created_at='2025-01-01' WHERE id=$1`, b)
	items, err = geoOffers(db).Discover(context.Background(), driver)
	if err != nil || items[0].RideRequestID != b {
		t.Fatalf("oldest tie should rank first: %+v %v", items, err)
	}
	geoExec(t, db, `UPDATE driver_locations SET latitude=0, longitude=0 WHERE driver_user_id=$1`, driver)
	geoExec(t, db, `UPDATE ride_requests SET pickup_latitude=0, pickup_longitude=180 WHERE id IN ($1,$2)`, a, b)
	items, err = geoOffers(db).Discover(context.Background(), driver)
	if err != nil || len(items) != 2 {
		t.Fatalf("no arbitrary radius cutoff: %+v %v", items, err)
	}
	if math.IsNaN(items[0].PickupDistanceMeters) || math.Abs(items[0].PickupDistanceMeters-math.Pi*6371000) > 1 {
		t.Fatalf("antipodal distance: %f", items[0].PickupDistanceMeters)
	}
}

func TestMarketplaceEligibilityIsConsistentAcrossTheJourney(t *testing.T) {
	for _, reason := range []string{"missing_location", "stale_location", "future_location", "offline", "missing_vehicle", "missing_capability", "busy"} {
		t.Run(reason, func(t *testing.T) {
			db := openCancellationIntegrationDB(t)
			ctx := context.Background()
			rider := createCancellationUser(t, db, "rider")
			driver := createCancellationDriver(t, db)
			ride := createCancellationRide(t, db, rider)
			offers := geoOffers(db)
			if _, err := offers.AcceptProposed(ctx, ride, driver); err != nil {
				t.Fatal(err)
			}
			switch reason {
			case "missing_location":
				geoExec(t, db, `DELETE FROM driver_locations WHERE driver_user_id=$1`, driver)
			case "stale_location":
				geoExec(t, db, `UPDATE driver_locations SET updated_at=NOW()-INTERVAL '121 seconds' WHERE driver_user_id=$1`, driver)
			case "future_location":
				geoExec(t, db, `UPDATE driver_locations SET updated_at=NOW()+INTERVAL '1 minute' WHERE driver_user_id=$1`, driver)
			case "offline":
				geoExec(t, db, `UPDATE driver_profiles SET is_online=FALSE WHERE user_id=$1`, driver)
			case "missing_vehicle":
				geoExec(t, db, `DELETE FROM driver_vehicles WHERE driver_user_id=$1`, driver)
			case "missing_capability":
				geoExec(t, db, `DELETE FROM user_capabilities WHERE user_id=$1 AND capability='driver'`, driver)
			case "busy":
				other := createCancellationRide(t, db, rider)
				insertCancellationTrip(t, db, other, rider, driver, trip.StatusAssigned)
			}
			items, err := offers.Discover(ctx, driver)
			if err != nil || len(items) != 0 {
				t.Fatalf("ineligible discovery: %+v %v", items, err)
			}
			view, err := offers.ListForRider(ctx, ride, rider)
			if err != nil || len(view) != 1 || view[0].Selectable || view[0].Status != offer.StatusPending {
				t.Fatalf("unavailability must not mutate pending offer: %+v %v", view, err)
			}
			if (reason == "missing_location" || reason == "stale_location" || reason == "future_location") && view[0].PickupDistanceMeters != nil {
				t.Fatal("unreliable location exposed as current distance")
			}
			if _, err := offers.Submit(ctx, ride, driver, 110000); !errors.Is(err, offer.ErrDriverIneligible) {
				t.Fatalf("submission bypassed eligibility: %v", err)
			}
			if _, err := offers.AcceptProposed(ctx, ride, driver); !errors.Is(err, offer.ErrDriverIneligible) {
				t.Fatalf("exact-fare bypassed eligibility: %v", err)
			}
			if _, err := offers.Accept(ctx, ride, rider, driver); !errors.Is(err, offer.ErrDriverIneligible) {
				t.Fatalf("selection bypassed eligibility: %v", err)
			}
			var count int
			if err := db.QueryRow(`SELECT count(*) FROM trips WHERE ride_request_id=$1`, ride).Scan(&count); err != nil || count != 0 {
				t.Fatalf("failed selection created Trip: %d %v", count, err)
			}
		})
	}
}

func TestRiderComparisonRefreshOwnershipAndHistory(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	stranger := createCancellationUser(t, db, "rider")
	near, far, stale := createCancellationDriver(t, db), createCancellationDriver(t, db), createCancellationDriver(t, db)
	ride := createCancellationRide(t, db, rider)
	geoExec(t, db, `UPDATE ride_requests SET pickup_latitude=0,pickup_longitude=0 WHERE id=$1`, ride)
	geoExec(t, db, `UPDATE driver_locations SET latitude=0,longitude=0 WHERE driver_user_id=$1`, near)
	geoExec(t, db, `UPDATE driver_locations SET latitude=0,longitude=1 WHERE driver_user_id=$1`, far)
	offers := geoOffers(db)
	for _, id := range []uuid.UUID{near, far} {
		if _, err := offers.AcceptProposed(ctx, ride, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := offers.Submit(ctx, ride, stale, 90000); err != nil {
		t.Fatal(err)
	}
	geoExec(t, db, `UPDATE driver_locations SET updated_at=NOW()-INTERVAL '3 minutes' WHERE driver_user_id=$1`, stale)
	view, err := offers.ListForRider(ctx, ride, rider)
	if err != nil {
		t.Fatal(err)
	}
	if len(view) != 3 || view[0].DriverUserID != near || view[1].DriverUserID != far || view[2].DriverUserID != stale {
		t.Fatalf("comparison ordering: %+v", view)
	}
	if !view[0].MatchesProposedFare || !view[0].Selectable || view[0].Vehicle == nil || view[0].Vehicle.Make != "Test" || view[0].Vehicle.Model != "Car" || view[0].Vehicle.Color != "White" || view[0].PickupDistanceMeters == nil || *view[0].PickupDistanceMeters != 0 {
		t.Fatalf("comparison details: %+v", view[0])
	}
	if view[2].MatchesProposedFare || view[2].PickupDistanceMeters != nil || view[2].Selectable {
		t.Fatalf("stale comparison: %+v", view[2])
	}
	if _, err := offers.ListForRider(ctx, ride, stranger); !errors.Is(err, offer.ErrRideNotFound) {
		t.Fatalf("ownership bypass: %v", err)
	}
	if _, err := offers.Accept(ctx, ride, stranger, near); !errors.Is(err, offer.ErrOfferNotActionable) {
		t.Fatalf("selection ownership bypass: %v", err)
	}
	locations := driverlocation.NewService(driverlocation.NewPostgresRepository(db))
	if _, err := locations.SetCurrent(ctx, stale, driverlocation.Input{Latitude: 0, Longitude: 0}); err != nil {
		t.Fatal(err)
	}
	view, err = offers.ListForRider(ctx, ride, rider)
	if err != nil || !view[0].Selectable || view[0].DriverUserID != stale {
		t.Fatalf("refresh did not restore pending offer: %+v %v", view, err)
	}
	if _, err := offers.Reject(ctx, ride, rider, near); err != nil {
		t.Fatal(err)
	}
	view, err = offers.ListForRider(ctx, ride, rider)
	if err != nil {
		t.Fatal(err)
	}
	if view[2].DriverUserID != near || view[2].Selectable || view[2].Status != offer.StatusRejected {
		t.Fatalf("rejected offer must remain unselectable history: %+v", view)
	}
	if _, err := offers.Accept(ctx, ride, rider, stale); err != nil {
		t.Fatalf("selection after refresh: %v", err)
	}
}
