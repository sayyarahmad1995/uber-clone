package cancellation

import (
	"context"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

func TestDriverPresentationRoundTripPreservesLegacyEligibility(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	riderID := createCancellationUser(t, db, "rider")
	driverID := createCancellationDriver(t, db)
	rideID := createCancellationRide(t, db, riderID)
	offers := geoOffers(db)
	assignments := marketplace.NewAssignmentService(marketplace.NewPostgresAssignmentRepository(db))
	submission, err := offers.AcceptProposed(ctx, rideID, driverID)
	if err != nil {
		t.Fatalf("legacy Driver response: %v", err)
	}
	view, err := offers.ListForRider(ctx, rideID, riderID)
	if err != nil {
		t.Fatal(err)
	}
	if len(view) != 1 || !view[0].Selectable || view[0].Driver != nil || view[0].Vehicle == nil || view[0].Vehicle.ModelYear != 0 {
		t.Fatalf("legacy comparison: %+v", view)
	}
	drivers := driver.NewService(driver.NewPostgresRepository(db))
	for _, name := range []string{" First Driver ", " Updated Driver "} {
		profile, err := drivers.Onboard(ctx, driverID, driver.OnboardingInput{DisplayName: name, Vehicle: driver.VehicleInput{Make: " Toyota ", Model: " Corolla ", ModelYear: 2024, Color: " White ", LicensePlate: " abc-123 "}})
		if err != nil {
			t.Fatal(err)
		}
		if !profile.IsOnline || profile.Vehicle.ModelYear != 2024 || profile.Vehicle.LicensePlate != "ABC-123" {
			t.Fatalf("re-onboarding changed operational state or lost fields: %+v", profile)
		}
		persisted, err := drivers.Get(ctx, driverID)
		if err != nil || persisted.DisplayName != profile.DisplayName {
			t.Fatalf("profile round trip: %+v %v", persisted, err)
		}
		// An already-submitted offer keeps its immutable public presentation snapshot.
		view, err = offers.ListForRider(ctx, rideID, riderID)
		if err != nil {
			t.Fatal(err)
		}
		if len(view) != 1 || !view[0].Selectable || view[0].Driver != nil || view[0].Vehicle == nil || view[0].Vehicle.ModelYear != 0 {
			t.Fatalf("submitted offer presentation snapshot changed: %+v", view)
		}
	}
	if _, err := assignments.SelectOffer(ctx, rideID, riderID, driverID, submission.Offer.UpdatedAt); err != nil {
		t.Fatalf("presentation update broke selection: %v", err)
	}
}
