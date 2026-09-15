package offer

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
)

func TestMarketplaceTimingPolicyIsSnapshottedAtCreation(t *testing.T) {
	db := openOfferIntegrationDB(t)
	ctx := context.Background()
	original, err := marketplace.LoadTimingPolicy(ctx, db)
	if err != nil {
		t.Fatalf("load original timing policy: %v", err)
	}
	t.Cleanup(func() {
		_, _ = marketplace.UpdateTimingPolicy(context.Background(), db, original, "test-restore")
	})

	firstPolicy := marketplace.TimingPolicy{
		RideRequestTTLSeconds:       240,
		DriverOpportunityTTLSeconds: 45,
		OfferDecisionTTLSeconds:     20,
	}
	if _, err := marketplace.UpdateTimingPolicy(ctx, db, firstPolicy, "timing-test"); err != nil {
		t.Fatalf("set first timing policy: %v", err)
	}

	riderID := createOfferTestUser(t, db, "rider")
	driverID := createOfferTestDriver(t, db)
	firstRide := createPolicyRide(t, db, riderID)
	assertDuration(t, "first ride TTL", firstRide.ExpiresAt.Sub(firstRide.CreatedAt), 240*time.Second)

	repository := NewPostgresRepository(db)
	feed, err := repository.Discover(ctx, driverID, DiscoveryLimit)
	if err != nil {
		t.Fatalf("discover first ride: %v", err)
	}
	firstDiscovery := findDiscovery(t, feed, firstRide.ID)
	var openedAt, visibleUntil time.Time
	if err := db.QueryRow(`
		SELECT opened_at, visible_until
		FROM driver_ride_request_opportunities
		WHERE ride_request_id=$1 AND driver_user_id=$2
	`, firstRide.ID, driverID).Scan(&openedAt, &visibleUntil); err != nil {
		t.Fatalf("read first opportunity: %v", err)
	}
	assertDuration(t, "first opportunity TTL", visibleUntil.Sub(openedAt), 45*time.Second)
	if !firstDiscovery.OpportunityExpiresAt.Equal(visibleUntil) {
		t.Fatal("discovery did not expose persisted opportunity deadline")
	}

	firstOffer, err := repository.Upsert(ctx, firstRide.ID, driverID, 100000, 90000, 130000, "PKR")
	if err != nil {
		t.Fatalf("submit first offer: %v", err)
	}
	assertDuration(t, "first offer TTL", firstOffer.ExpiresAt.Sub(firstOffer.CreatedAt), 20*time.Second)

	firstRideExpiry := firstRide.ExpiresAt
	firstOpportunityExpiry := visibleUntil
	firstOfferExpiry := firstOffer.ExpiresAt

	secondPolicy := marketplace.TimingPolicy{
		RideRequestTTLSeconds:       300,
		DriverOpportunityTTLSeconds: 50,
		OfferDecisionTTLSeconds:     25,
	}
	if _, err := marketplace.UpdateTimingPolicy(ctx, db, secondPolicy, "timing-test-2"); err != nil {
		t.Fatalf("set second timing policy: %v", err)
	}

	var persistedRideExpiry, persistedOpportunityExpiry, persistedOfferExpiry time.Time
	if err := db.QueryRow(`SELECT expires_at FROM ride_requests WHERE id=$1`, firstRide.ID).Scan(&persistedRideExpiry); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT visible_until FROM driver_ride_request_opportunities WHERE ride_request_id=$1 AND driver_user_id=$2`, firstRide.ID, driverID).Scan(&persistedOpportunityExpiry); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT expires_at FROM ride_offers WHERE ride_request_id=$1 AND driver_user_id=$2`, firstRide.ID, driverID).Scan(&persistedOfferExpiry); err != nil {
		t.Fatal(err)
	}
	if !persistedRideExpiry.Equal(firstRideExpiry) || !persistedOpportunityExpiry.Equal(firstOpportunityExpiry) || !persistedOfferExpiry.Equal(firstOfferExpiry) {
		t.Fatal("policy update retroactively changed an existing marketplace deadline")
	}

	secondDriverID := createOfferTestDriver(t, db)
	secondRide := createPolicyRide(t, db, riderID)
	assertDuration(t, "second ride TTL", secondRide.ExpiresAt.Sub(secondRide.CreatedAt), 300*time.Second)
	feed, err = repository.Discover(ctx, secondDriverID, DiscoveryLimit)
	if err != nil {
		t.Fatalf("discover second ride: %v", err)
	}
	findDiscovery(t, feed, secondRide.ID)
	if err := db.QueryRow(`
		SELECT opened_at, visible_until
		FROM driver_ride_request_opportunities
		WHERE ride_request_id=$1 AND driver_user_id=$2
	`, secondRide.ID, secondDriverID).Scan(&openedAt, &visibleUntil); err != nil {
		t.Fatalf("read second opportunity: %v", err)
	}
	assertDuration(t, "second opportunity TTL", visibleUntil.Sub(openedAt), 50*time.Second)
	secondOffer, err := repository.Upsert(ctx, secondRide.ID, secondDriverID, 100000, 90000, 130000, "PKR")
	if err != nil {
		t.Fatalf("submit second offer: %v", err)
	}
	assertDuration(t, "second offer TTL", secondOffer.ExpiresAt.Sub(secondOffer.CreatedAt), 25*time.Second)
}

func createPolicyRide(t *testing.T, db interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, riderID uuid.UUID) ride.Request {
	t.Helper()
	// Kept below as a concrete repository call so the production creation path
	// snapshots the current timing policy.
	database := db.(*sql.DB)
	request, err := ride.NewService(ride.NewPostgresRepository(database)).Create(context.Background(), riderID, ride.CreateInput{
		ServiceCode: "economy",
		Pickup:      ride.Location{Latitude: 24.8610, Longitude: 67.0010},
		Destination: ride.Location{Latitude: 24.8800, Longitude: 67.0200},
		ProposedFare: &ride.Money{
			AmountMinor: 100000,
			Currency:    "PKR",
		},
	})
	if err != nil {
		t.Fatalf("create policy ride: %v", err)
	}
	t.Cleanup(func() {
		_, _ = database.Exec(`DELETE FROM trips WHERE ride_request_id=$1`, request.ID)
		_, _ = database.Exec(`DELETE FROM ride_offers WHERE ride_request_id=$1`, request.ID)
		_, _ = database.Exec(`DELETE FROM driver_ride_request_opportunities WHERE ride_request_id=$1`, request.ID)
		_, _ = database.Exec(`DELETE FROM ride_requests WHERE id=$1`, request.ID)
	})
	return request
}

func findDiscovery(t *testing.T, items []DiscoveryItem, rideID uuid.UUID) DiscoveryItem {
	t.Helper()
	for _, item := range items {
		if item.RideRequestID == rideID {
			return item
		}
	}
	t.Fatalf("ride %s not found in Driver discovery", rideID)
	return DiscoveryItem{}
}

func assertDuration(t *testing.T, name string, got, want time.Duration) {
	t.Helper()
	if delta := got - want; delta < -time.Second || delta > time.Second {
		t.Fatalf("%s=%v want approximately %v", name, got, want)
	}
}
