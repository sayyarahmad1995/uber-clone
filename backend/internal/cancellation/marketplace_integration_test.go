package cancellation

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
)

func TestMarketplaceWithoutCandidatesThroughCancellation(t *testing.T) {
	db := openCancellationIntegrationDB(t)
	ctx := context.Background()
	rider := createCancellationUser(t, db, "rider")
	driver := createCancellationDriver(t, db)
	competitor := createCancellationDriver(t, db)
	rides := ride.NewService(ride.NewPostgresRepository(db))
	create := func() uuid.UUID {
		r, err := rides.Create(ctx, rider, ride.CreateInput{
			Pickup:       ride.Location{Latitude: 24, Longitude: 67},
			Destination:  ride.Location{Latitude: 25, Longitude: 68},
			ProposedFare: &ride.Money{AmountMinor: 100000, Currency: "PKR"},
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			db.Exec(`DELETE FROM trips WHERE ride_request_id=$1`, r.ID)
			db.Exec(`DELETE FROM ride_offers WHERE ride_request_id=$1`, r.ID)
			db.Exec(`DELETE FROM driver_ride_request_opportunities WHERE ride_request_id=$1`, r.ID)
			db.Exec(`DELETE FROM ride_requests WHERE id=$1`, r.ID)
		})
		return r.ID
	}
	first, second, legacy := create(), create(), create()
	if _, err := db.Exec(`UPDATE ride_requests SET proposed_fare_minor=NULL, currency=NULL WHERE id=$1`, legacy); err != nil {
		t.Fatal(err)
	}
	offers := offer.NewService(offer.NewPostgresRepository(db))
	assignments := marketplace.NewAssignmentService(marketplace.NewPostgresAssignmentRepository(db))
	feed, err := offers.Discover(ctx, driver)
	if err != nil {
		t.Fatal(err)
	}
	found := map[uuid.UUID]bool{}
	for _, item := range feed {
		found[item.RideRequestID] = true
	}
	if !found[first] || !found[second] || found[legacy] {
		t.Fatalf("unexpected discovery: %v", found)
	}
	submittedOffers := make(map[uuid.UUID]offer.Offer)
	for _, id := range []uuid.UUID{first, second} {
		submission, err := offers.AcceptProposed(ctx, id, driver)
		if err != nil {
			t.Fatalf("accept Rider fare: %v", err)
		}
		if submission.Offer.Status != offer.StatusPending || submission.Offer.AmountMinor != 100000 {
			t.Fatalf("expected pending exact-fare offer: %+v", submission)
		}
		submittedOffers[id] = submission.Offer
	}
	competitorFeed, err := offers.Discover(ctx, competitor)
	if err != nil {
		t.Fatal(err)
	}
	competitorFound := false
	for _, item := range competitorFeed {
		if item.RideRequestID == first {
			competitorFound = true
		}
	}
	if !competitorFound {
		t.Fatal("competitor did not receive a one-time opportunity for first ride")
	}
	if _, err := offers.Submit(ctx, first, competitor, 110000); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM trips WHERE ride_request_id IN ($1,$2)`, first, second).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("Driver response assigned a Trip before Rider selection")
	}
	if _, err := assignments.SelectOffer(ctx, first, rider, driver, submittedOffers[first].UpdatedAt); err != nil {
		t.Fatal(err)
	}
	feed, err = offers.Discover(ctx, driver)
	if err != nil || len(feed) != 0 {
		t.Fatalf("busy Driver should not discover requests: %v %v", feed, err)
	}
	if _, err := offers.AcceptProposed(ctx, second, driver); !errors.Is(err, offer.ErrOpportunityNotOpen) {
		t.Fatalf("one-shot Driver submission was reopened while busy: %v", err)
	}
	if _, err := assignments.SelectOffer(ctx, second, rider, driver, submittedOffers[second].UpdatedAt); !errors.Is(err, marketplace.ErrDriverUnavailable) {
		t.Fatalf("busy Driver selection: %v", err)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM ride_offers WHERE ride_request_id=$1 AND driver_user_id=$2`, first, competitor).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "closed" {
		t.Fatalf("competing offer remains %s", status)
	}
	if _, err := NewPostgresRepository(db).CancelByRider(ctx, first, rider); err != nil {
		t.Fatal(err)
	}
	feed, err = offers.Discover(ctx, driver)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range feed {
		if item.RideRequestID == first || item.RideRequestID == second || item.RideRequestID == legacy {
			t.Fatalf("terminal or already-responded request reappeared after cancellation: %+v", feed)
		}
	}
	comparison, err := offers.ListForRider(ctx, second, rider)
	if err != nil {
		t.Fatal(err)
	}
	if len(comparison) != 1 || !comparison[0].Selectable || comparison[0].DriverUserID != driver {
		t.Fatalf("existing second offer did not become selectable after Driver was freed: %+v", comparison)
	}
	if _, err := assignments.SelectOffer(ctx, second, rider, driver, submittedOffers[second].UpdatedAt); err != nil {
		t.Fatalf("freed Driver selection: %v", err)
	}
}
