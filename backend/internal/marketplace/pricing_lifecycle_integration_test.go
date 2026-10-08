package marketplace_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/cancellation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
	"net/url"
	"os"
	"strings"
	"testing"
)

func pricingLifecycleDB(t *testing.T) *sql.DB {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("dedicated test database required")
	}
	db, err := database.Open(raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Apply(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPricingChangesDoNotRepriceRide(t *testing.T) {
	db := pricingLifecycleDB(t)
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	rider, driver, vehicle := uuid.New(), uuid.New(), uuid.New()
	exec("INSERT INTO users(id) VALUES($1),($2)", rider, driver)
	exec("INSERT INTO user_capabilities(user_id,capability) VALUES($1,'rider'),($2,'driver')", rider, driver)
	exec("INSERT INTO driver_profiles(user_id,status,is_online) VALUES($1,'active',true)", driver)
	exec("INSERT INTO driver_vehicles(id,driver_user_id,make,model,color,license_plate) VALUES($1,$2,'Test','Car','White','PRICING')", vehicle, driver)
	exec("INSERT INTO driver_vehicle_service_enrollments(vehicle_id,service_code,approved_at,approved_by) VALUES($1,'economy',NOW(),'pricing-test')", vehicle)
	exec("INSERT INTO driver_operating_selections(driver_user_id,vehicle_id) VALUES($1,$2)", driver, vehicle)
	exec("INSERT INTO driver_locations(driver_user_id,latitude,longitude,updated_at) VALUES($1,24.86,67.0,NOW())", driver)
	policies := pricing.NewPostgresRepository(db)
	draft := pricing.Draft{ServiceCode: "economy", Currency: "PKR", BaseFareMinor: 10000, RateMinorPerKm: 1000, RateMinorPerMinute: 100, MinimumFareMinor: 10000, RoundingIncrementMinor: 100}
	publish := func(base int64) pricing.Policy {
		t.Helper()
		draft.BaseFareMinor = base
		p, err := policies.Publish(ctx, draft, "pricing-test")
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	a := publish(10000)
	input := ride.CreateInput{ServiceCode: "economy", Pickup: ride.Location{Latitude: 24.861, Longitude: 67.001}, Destination: ride.Location{Latitude: 24.88, Longitude: 67.02}, ProposedFare: &ride.Money{AmountMinor: 110000, Currency: "PKR"}, PricingPolicyVersion: a.ID.String()}
	requests := ride.NewService(ride.NewPostgresRepositoryWithPricing(db, policies, true))
	request, err := requests.Create(ctx, rider, input)
	if err != nil {
		t.Fatal(err)
	}
	b := publish(900000)
	offers := offer.NewService(offer.NewPostgresRepository(db))
	market, err := offer.NewPostgresRepository(db).Market(ctx, request.ID)
	if err != nil {
		t.Fatal(err)
	}
	min, max := offer.Bounds(market.ProposedAmountMinor)
	if min != 99000 || max != 143000 {
		t.Fatalf("bounds repriced: %d..%d", min, max)
	}
	if _, err := offers.Submit(ctx, request.ID, driver, 98999); !errors.Is(err, offer.ErrAmountOutOfRange) {
		t.Fatalf("lower bound: %v", err)
	}
	if _, err := offers.Submit(ctx, request.ID, driver, 143001); !errors.Is(err, offer.ErrAmountOutOfRange) {
		t.Fatalf("upper bound: %v", err)
	}
	exec("INSERT INTO driver_ride_request_opportunities(ride_request_id,driver_user_id,status,visible_until) VALUES($1,$2,'open',NOW()+INTERVAL '60 seconds')", request.ID, driver)
	submitted, err := offers.Submit(ctx, request.ID, driver, 120000)
	if err != nil {
		t.Fatal(err)
	}
	assigned, err := marketplace.NewPostgresAssignmentRepository(db).SelectOffer(ctx, request.ID, rider, driver, submitted.Offer.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.OperationContext == nil || assigned.OperationContext.Fare.AmountMinor != 120000 {
		t.Fatalf("agreed fare: %+v", assigned)
	}
	c := publish(700000)
	exec("UPDATE driver_service_catalog SET is_active=false WHERE code='economy'")
	t.Cleanup(func() { _, _ = db.Exec("UPDATE driver_service_catalog SET is_active=true WHERE code='economy'") })
	// A disabled pricing pointer must not strand an already assigned Trip.
	if err := policies.Disable(ctx, "economy", "PKR", c.ID, "pricing-test"); err != nil {
		t.Fatal(err)
	}
	trips := trip.NewService(trip.NewPostgresRepository(db))
	if _, err := trips.Start(ctx, request.ID, driver); err != nil {
		t.Fatal(err)
	}
	if _, err := trips.Complete(ctx, request.ID, driver); err != nil {
		t.Fatal(err)
	}
	recovery := ridestatus.NewPostgresRepository(db)
	view, err := recovery.GetOwned(ctx, request.ID, rider)
	if err != nil {
		t.Fatal(err)
	}
	if view.Trip == nil || view.Trip.Status != trip.StatusCompleted || view.Trip.Settlement.Status != trip.SettlementUnsettled || view.Trip.OperationContext.Fare.AmountMinor != 120000 {
		t.Fatalf("unsettled recovery: %+v", view)
	}
	settled, err := trips.ConfirmCashCollected(ctx, request.ID, driver)
	if err != nil {
		t.Fatal(err)
	}
	if settled.Settlement.Status != trip.SettlementCashCollected || settled.OperationContext.Fare.AmountMinor != 120000 {
		t.Fatalf("settled fare: %+v", settled)
	}
	view, err = recovery.GetOwned(ctx, request.ID, rider)
	if err != nil {
		t.Fatal(err)
	}
	if view.RideRequest.ProposedFare.AmountMinor != 110000 || view.Trip.OperationContext.Fare.AmountMinor != 120000 || view.Trip.Settlement.Status != trip.SettlementCashCollected {
		t.Fatalf("recovered amounts changed: %+v", view)
	}
	var policyID uuid.UUID
	var base, version int64
	if err := db.QueryRow("SELECT policy_id,base_fare_minor,version FROM ride_request_pricing_snapshots WHERE ride_request_id=$1", request.ID).Scan(&policyID, &base, &version); err != nil {
		t.Fatal(err)
	}
	if policyID != a.ID || base != a.BaseFareMinor || version != a.Version || policyID == b.ID {
		t.Fatal("original snapshot changed")
	}
	exec("UPDATE driver_service_catalog SET is_active=true WHERE code='economy'")
	// Legacy manual requests remain usable while pricing is disabled.
	legacy, err := ride.NewService(ride.NewPostgresRepository(db)).Create(ctx, rider, input)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.ProposedFare.AmountMinor != 110000 || legacy.PricingSnapshot != nil {
		t.Fatal("legacy request repriced")
	}
	exec("INSERT INTO driver_ride_request_opportunities(ride_request_id,driver_user_id,status,visible_until) VALUES($1,$2,'open',NOW()+INTERVAL '60 seconds')", legacy.ID, driver)
	legacyOffer, err := offers.Submit(ctx, legacy.ID, driver, 120000)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := marketplace.NewPostgresAssignmentRepository(db).SelectOffer(ctx, legacy.ID, rider, driver, legacyOffer.Offer.UpdatedAt); err != nil {
		t.Fatal(err)
	}
	if _, err := trips.Start(ctx, legacy.ID, driver); err != nil {
		t.Fatal(err)
	}
	if _, err := trips.Complete(ctx, legacy.ID, driver); err != nil {
		t.Fatal(err)
	}
	if _, err := trips.ConfirmCashCollected(ctx, legacy.ID, driver); err != nil {
		t.Fatal(err)
	}
	legacyView, err := recovery.GetOwned(ctx, legacy.ID, rider)
	if err != nil {
		t.Fatal(err)
	}
	if legacyView.RideRequest.ProposedFare.AmountMinor != 110000 || legacyView.Trip.OperationContext.Fare.AmountMinor != 120000 || legacyView.Trip.Settlement.Status != trip.SettlementCashCollected {
		t.Fatal("legacy recovery repriced")
	}
	legacyCancel, err := ride.NewService(ride.NewPostgresRepository(db)).Create(ctx, rider, input)
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := cancellation.NewPostgresRepository(db).CancelByRider(ctx, legacyCancel.ID, rider)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != ride.StatusCancelled {
		t.Fatal("legacy cancellation failed")
	}
	latest := publish(600000)
	input.PricingPolicyVersion = latest.ID.String()
	cancelledRequest, err := requests.Create(ctx, rider, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cancellation.NewPostgresRepository(db).CancelByRider(ctx, cancelledRequest.ID, rider); err != nil {
		t.Fatal(err)
	}
	if err := policies.Disable(ctx, "economy", "PKR", latest.ID, "pricing-test"); err != nil {
		t.Fatal(err)
	}
	var copied uuid.UUID
	if err := db.QueryRow("SELECT policy_id FROM ride_request_pricing_snapshots WHERE ride_request_id=$1", cancelledRequest.ID).Scan(&copied); err != nil || copied != latest.ID {
		t.Fatalf("cancelled snapshot lost: %v", err)
	}
	if _, err := requests.Create(ctx, rider, input); !errors.Is(err, pricing.ErrUnavailable) {
		t.Fatalf("disabled creation: %v", err)
	}
}
