package cancellation

import (
 "context"
 "encoding/json"
 "errors"
 "testing"

 "github.com/google/uuid"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
)

func TestMarketplaceRequiresMatchingReviewedOperationAndPreservesTrip(t *testing.T) {
 db := openCancellationIntegrationDB(t)
 ctx := context.Background()
 rider := createCancellationUser(t,db,"rider")
 driverID := createCancellationDriver(t,db)
 rideID := createCancellationRide(t,db,rider)
 offers := geoOffers(db)
 mustExec := func(query string, args ...any) { t.Helper(); if _,err := db.Exec(query,args...); err != nil { t.Fatal(err) } }
 mustExec(`UPDATE ride_requests SET service_code='comfort' WHERE id=$1`,rideID)
 feed,err := offers.Discover(ctx,driverID)
 if err != nil { t.Fatal(err) }
 for _,item := range feed { if item.RideRequestID==rideID { t.Fatal("economy Driver discovered Comfort request") } }
 if _,err := offers.AcceptProposed(ctx,rideID,driverID); !errors.Is(err,offer.ErrDriverIneligible) { t.Fatalf("wrong service offer: %v",err) }
 mustExec(`UPDATE ride_requests SET service_code='economy' WHERE id=$1`,rideID)
 if _,err := offers.AcceptProposed(ctx,rideID,driverID); err != nil { t.Fatal(err) }
 second := uuid.New()
 mustExec(`INSERT INTO driver_vehicles (id,driver_user_id,make,model,model_year,color,license_plate) VALUES ($1,$2,'Toyota','Second',2024,'Blue','SECOND')`,second,driverID)
 mustExec(`INSERT INTO driver_vehicle_service_enrollments (vehicle_id,service_code,approved_at,approved_by) VALUES ($1,'economy',NOW(),'reviewer')`,second)
 mustExec(`UPDATE driver_operating_selections SET vehicle_id=$1 WHERE driver_user_id=$2`,second,driverID)
 comparison,err := offers.ListForRider(ctx,rideID,rider)
 if err != nil || len(comparison)!=1 || comparison[0].Selectable || comparison[0].Vehicle.Model!="Car" { t.Fatalf("changed operation selectable or snapshot lost: %+v %v",comparison,err) }
 if _,err := offers.Accept(ctx,rideID,rider,driverID); !errors.Is(err,offer.ErrDriverIneligible) { t.Fatalf("stale operation assigned: %v",err) }
 if _,err := offers.Submit(ctx,rideID,driverID,110000); err != nil { t.Fatal(err) }
 assigned,err := offers.Accept(ctx,rideID,rider,driverID)
 if err != nil { t.Fatal(err) }
 var snapshot map[string]any
 if err := json.Unmarshal(assigned.OperationContext,&snapshot); err != nil { t.Fatal(err) }
 if snapshot["vehicle_id"]!=second.String() || snapshot["model"]!="Second" || snapshot["service_code"]!="economy" || snapshot["fare"].(map[string]any)["amount_minor"]!=float64(110000) { t.Fatalf("wrong assignment context: %s",assigned.OperationContext) }
 trips := trip.NewPostgresRepository(db)
 if _,err := trips.Start(ctx,rideID,driverID); err != nil { t.Fatal(err) }
 if _,err := trips.Complete(ctx,rideID,driverID); err != nil { t.Fatal(err) }
 mustExec(`UPDATE driver_vehicles SET model='Changed later' WHERE id=$1`,second)
 mustExec(`DELETE FROM driver_vehicle_service_enrollments WHERE vehicle_id=$1`,second)
 view,err := ridestatus.NewPostgresRepository(db).GetOwned(ctx,rideID,rider)
 if err != nil || view.Trip == nil || string(view.Trip.OperationContext)!=string(assigned.OperationContext) { t.Fatalf("Rider history changed: %+v %v",view,err) }
 history,err := drivertrip.NewPostgresRepository(db).ListHistory(ctx,driverID,50)
 if err != nil || len(history)!=1 || string(history[0].OperationContext)!=string(assigned.OperationContext) { t.Fatalf("Driver history changed: %+v %v",history,err) }
}

func TestMarketplaceFailsClosedWithoutEnrollment(t *testing.T) {
 db := openCancellationIntegrationDB(t)
 ctx := context.Background()
 rider := createCancellationUser(t,db,"rider")
 driverID := createCancellationDriver(t,db)
 rideID := createCancellationRide(t,db,rider)
 offers := geoOffers(db)
 if _,err := offers.AcceptProposed(ctx,rideID,driverID); err != nil { t.Fatal(err) }
 if _,err := db.Exec(`DELETE FROM driver_vehicle_service_enrollments WHERE vehicle_id IN (SELECT id FROM driver_vehicles WHERE driver_user_id=$1)`,driverID); err != nil { t.Fatal(err) }
 if _,err := offers.Accept(ctx,rideID,rider,driverID); !errors.Is(err,offer.ErrDriverIneligible) { t.Fatalf("revoked approval assigned: %v",err) }
 if _,err := offers.AcceptProposed(ctx,rideID,driverID); !errors.Is(err,offer.ErrDriverIneligible) { t.Fatalf("legacy driver offered: %v",err) }
}

func TestRiderMustAcceptTheDisplayedOfferRevision(t *testing.T) {
 db := openCancellationIntegrationDB(t)
 ctx := context.Background()
 rider := createCancellationUser(t,db,"rider")
 driverID := createCancellationDriver(t,db)
 rideID := createCancellationRide(t,db,rider)
 offers := geoOffers(db)
 first,err := offers.AcceptProposed(ctx,rideID,driverID)
 if err != nil { t.Fatal(err) }
 second,err := offers.Submit(ctx,rideID,driverID,120000)
 if err != nil { t.Fatal(err) }
 if _,err := offers.Accept(ctx,rideID,rider,driverID,first.Offer.UpdatedAt); !errors.Is(err,offer.ErrOfferNotActionable) { t.Fatalf("accepted a revised fare: %v",err) }
 if _,err := offers.Accept(ctx,rideID,rider,driverID,second.Offer.UpdatedAt); err != nil { t.Fatal(err) }
}
