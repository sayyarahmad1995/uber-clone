package cancellation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

func TestMarketplaceRequiresMatchingReviewedOperationAndPreservesTrip(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, rider)
	offers := geoOffers(db)
	assignments := marketplace.NewAssignmentService(marketplace.NewPostgresAssignmentRepository(db))
	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(`UPDATE ride_requests SET service_code='comfort' WHERE id=$1`, rideID)
	feed, err := offers.Discover(ctx, driverID)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range feed {
		if item.RideRequestID == rideID {
			t.Fatal("economy Driver discovered Comfort request")
		}
	}
	if _, err := offers.AcceptProposed(ctx, rideID, driverID); !errors.Is(err, offer.ErrDriverIneligible) {
		t.Fatalf("wrong service offer: %v", err)
	}
	mustExec(`UPDATE ride_requests SET service_code='economy' WHERE id=$1`, rideID)
	submission, err := offers.AcceptProposed(ctx, rideID, driverID)
	if err != nil {
		t.Fatal(err)
	}
	second := uuid.New()
	mustExec(`INSERT INTO driver_vehicles (id,driver_user_id,make,model,model_year,color,license_plate) VALUES ($1,$2,'Toyota','Second',2024,'Blue','SECOND')`, second, driverID)
	mustExec(`INSERT INTO driver_vehicle_service_enrollments (vehicle_id,service_code,approved_at,approved_by) VALUES ($1,'economy',NOW(),'reviewer')`, second)
	mustExec(`UPDATE driver_operating_selections SET vehicle_id=$1 WHERE driver_user_id=$2`, second, driverID)
	comparison, err := offers.ListForRider(ctx, rideID, rider)
	if err != nil || len(comparison) != 1 || comparison[0].Selectable || comparison[0].Vehicle.Model != "Car" {
		t.Fatalf("changed operation selectable or snapshot lost: %+v %v", comparison, err)
	}
	if _, err := assignments.SelectOffer(ctx, rideID, rider, driverID, submission.Offer.UpdatedAt); !errors.Is(err, marketplace.ErrDriverUnavailable) {
		t.Fatalf("stale operation assigned: %v", err)
	}
	if _, err := offers.Submit(ctx, rideID, driverID, 110000); !errors.Is(err, offer.ErrOpportunityNotOpen) {
		t.Fatalf("same Driver re-offered after operation changed: %v", err)
	}

	// A new ride can snapshot the newly selected vehicle/service operation.
	freshRideID := createCancellationRide(t, db, rider)
	freshSubmission, err := offers.Submit(ctx, freshRideID, driverID, 110000)
	if err != nil {
		t.Fatalf("fresh ride offer: %v", err)
	}
	assigned, err := assignments.SelectOffer(ctx, freshRideID, rider, driverID, freshSubmission.Offer.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.OperationContext == nil || assigned.OperationContext.VehicleID != second || assigned.OperationContext.Model != "Second" || assigned.OperationContext.ServiceCode != "economy" || assigned.OperationContext.Fare.AmountMinor != 110000 {
		t.Fatalf("wrong assignment context: %+v", assigned.OperationContext)
	}
	trips := trip.NewPostgresRepository(db)
	if _, err := trips.Start(ctx, freshRideID, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err := trips.Complete(ctx, freshRideID, driverID); err != nil {
		t.Fatal(err)
	}
	mustExec(`UPDATE driver_vehicles SET model='Changed later' WHERE id=$1`, second)
	mustExec(`DELETE FROM driver_vehicle_service_enrollments WHERE vehicle_id=$1`, second)
	view, err := ridestatus.NewPostgresRepository(db).GetOwned(ctx, freshRideID, rider)
	if err != nil || view.Trip == nil || view.Trip.OperationContext == nil || *view.Trip.OperationContext != *assigned.OperationContext {
		t.Fatalf("Rider history changed: %+v %v", view, err)
	}
	history, err := drivertrip.NewPostgresRepository(db).ListHistory(ctx, driverID, 50)
	if err != nil || len(history) != 1 || history[0].OperationContext == nil || *history[0].OperationContext != *assigned.OperationContext {
		t.Fatalf("Driver history changed: %+v %v", history, err)
	}
}

func TestMarketplaceFailsClosedWithoutEnrollment(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, rider)
	offers := geoOffers(db)
	assignments := marketplace.NewAssignmentService(marketplace.NewPostgresAssignmentRepository(db))
	submission, err := offers.AcceptProposed(ctx, rideID, driverID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM driver_vehicle_service_enrollments WHERE vehicle_id IN (SELECT id FROM driver_vehicles WHERE driver_user_id=$1)`, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err := assignments.SelectOffer(ctx, rideID, rider, driverID, submission.Offer.UpdatedAt); !errors.Is(err, marketplace.ErrDriverUnavailable) {
		t.Fatalf("revoked approval assigned: %v", err)
	}
	if _, err := offers.AcceptProposed(ctx, rideID, driverID); !errors.Is(err, offer.ErrOpportunityNotOpen) {
		t.Fatalf("one-shot Driver offer reopened after approval revocation: %v", err)
	}
}

func TestDriverCannotReviseSubmittedOffer(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, rider)
	offers := geoOffers(db)
	first, err := offers.AcceptProposed(ctx, rideID, driverID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := offers.Submit(ctx, rideID, driverID, 120000); !errors.Is(err, offer.ErrOpportunityNotOpen) {
		t.Fatalf("submitted offer was revised: %v", err)
	}
	assignments := marketplace.NewAssignmentService(marketplace.NewPostgresAssignmentRepository(db))
	if _, err := assignments.SelectOffer(ctx, rideID, rider, driverID, first.Offer.UpdatedAt); err != nil {
		t.Fatalf("Rider could not accept original displayed offer: %v", err)
	}
}

func TestRiderReadsUnassignedRequestWithAndWithoutOffer(t *testing.T) {
	for _, withOffer := range []bool{false, true} {
		name := "before_offer"
		if withOffer {
			name = "after_driver_offer"
		}
		t.Run(name, func(t *testing.T) {
			db := openCancellationIntegrationDB(t)
			ctx := context.Background()
			rider := createCancellationUser(t, db, "rider")
			driverID := createCancellationDriver(t, db)
			rideID := createCancellationRide(t, db, rider)
			offers := geoOffers(db)
			if withOffer {
				if _, err := offers.Submit(ctx, rideID, driverID, 110000); err != nil {
					t.Fatal(err)
				}
			}
			statuses := ridestatus.NewPostgresRepository(db)
			view, err := statuses.GetOwned(ctx, rideID, rider)
			if err != nil {
				t.Fatalf("get unassigned request: %v", err)
			}
			if view.Trip != nil {
				t.Fatal("an offer must not assign a trip")
			}
			views, err := statuses.ListOwned(ctx, rider, 50)
			if err != nil || len(views) != 1 {
				t.Fatalf("list unassigned requests: %v, count=%d", err, len(views))
			}
			comparison, err := offers.ListForRider(ctx, rideID, rider)
			if err != nil {
				t.Fatal(err)
			}
			if withOffer && (len(comparison) != 1 || comparison[0].AmountMinor != 110000 || !comparison[0].Selectable) {
				t.Fatalf("counteroffer missing or not selectable: %+v", comparison)
			}
		})
	}
}

func TestLegacyTripWithNullContextRemainsReadableAndExecutable(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, rider)
	offers := geoOffers(db)
	assignments := marketplace.NewAssignmentService(marketplace.NewPostgresAssignmentRepository(db))
	submission, err := offers.AcceptProposed(ctx, rideID, driverID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := assignments.SelectOffer(ctx, rideID, rider, driverID, submission.Offer.UpdatedAt); err != nil {
		t.Fatal(err)
	}
	// Pre-migration trips have no captured context. Never fabricate one on read.
	if _, err := db.Exec(`UPDATE trips SET operation_context=NULL WHERE ride_request_id=$1`, rideID); err != nil {
		t.Fatal(err)
	}
	t.Run("rider_read", func(t *testing.T) {
		if _, err := ridestatus.NewPostgresRepository(db).GetOwned(ctx, rideID, rider); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("driver_read", func(t *testing.T) {
		if _, err := drivertrip.NewPostgresRepository(db).GetCurrent(ctx, driverID); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("execution_and_history", func(t *testing.T) {
		trips := trip.NewPostgresRepository(db)
		if _, err := trips.Start(ctx, rideID, driverID); err != nil {
			t.Fatal(err)
		}
		if _, err := trips.Complete(ctx, rideID, driverID); err != nil {
			t.Fatal(err)
		}
		history, err := drivertrip.NewPostgresRepository(db).ListHistory(ctx, driverID, 50)
		if err != nil || len(history) != 1 {
			t.Fatalf("legacy history: %v count=%d", err, len(history))
		}
	})
}
