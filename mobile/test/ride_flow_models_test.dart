import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_offer.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_snapshot.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';

void main() {
  test('parses pending driver offer', () {
    final offer = RideOffer.fromJson({'ride_request_id':'ride-1','driver_user_id':'driver-1','fare':{'amount_minor':11500,'currency':'PKR'},'status':'pending','created_at':'2026-09-15T10:00:00Z','updated_at':'2026-09-15T10:01:00Z','expires_at':'2026-09-15T10:02:00Z','decided_at':null});
    expect(offer.status, 'pending');
    expect(offer.fare.amountMinor, 11500);
    expect(offer.decidedAt, isNull);
  });

  test('parses selectable Rider offer summaries', () {
    final comparison = RiderOfferComparison.fromJson({'ride_request_id':'ride-1','driver_user_id':'driver-1','fare':{'amount_minor':10000,'currency':'PKR'},'status':'pending','created_at':'2026-09-15T10:00:00Z','updated_at':'2026-09-15T10:01:00Z','expires_at':'2026-09-15T10:02:00Z','decided_at':null,'pickup_distance_meters':840,'matches_proposed_fare':true,'selectable':true,'driver':{'display_name':'Ayesha'},'vehicle':{'make':'Toyota','model':'Corolla','model_year':2024,'color':'White'},'service':{'code':'car','display_name':'Car'}});
    expect(comparison.selectable, isTrue);
    expect(comparison.driver?.displayName, 'Ayesha');
    expect(comparison.vehicle?.modelYear, 2024);
    expect(comparison.service?.code, 'car');
  });

  test('parses assigned Rider ride snapshot', () {
    final ride = RiderRideSnapshot.fromJson({'service_code':'car','id':'ride-1','pickup':{'latitude':33.6844,'longitude':73.0479},'destination':{'latitude':33.5651,'longitude':73.0169},'proposed_fare':{'amount_minor':10000,'currency':'PKR'},'status':'accepted','created_at':'2026-09-15T10:00:00Z','expires_at':'2026-09-15T10:10:00Z','trip':{'operation_context':null,'driver_user_id':'driver-1','status':'assigned','assigned_at':'2026-09-15T10:02:00Z','started_at':null,'completed_at':null,'cancelled_at':null,'settlement':{'status':'unsettled','method':null,'cash_collected_at':null}}});
    expect(ride.trip?.status, 'assigned');
    expect(ride.trip?.driverUserId, 'driver-1');
    expect(ride.trip?.settlement.status, 'unsettled');
  });

  test('keeps completed Trip and cash settlement independent', () {
    final trip = TripSnapshot.fromJson({'operation_context':{'vehicle_id':'vehicle-1','service_code':'car','service_name':'Car','driver_name':'Ayesha','make':'Toyota','model':'Corolla','model_year':2024,'color':'White','license_plate':'ABC-123','fare':{'amount_minor':10000,'currency':'PKR'}},'ride_request_id':'ride-1','pickup':{'latitude':33.6844,'longitude':73.0479},'destination':{'latitude':33.5651,'longitude':73.0169},'status':'completed','assigned_at':'2026-09-15T10:02:00Z','started_at':'2026-09-15T10:05:00Z','completed_at':'2026-09-15T10:25:00Z','cancelled_at':null,'settlement':{'status':'cash_collected','method':'cash','cash_collected_at':'2026-09-15T10:26:00Z'}});
    expect(trip.status, 'completed');
    expect(trip.settlement.status, 'cash_collected');
    expect(trip.operationContext?.fare.amountMinor, 10000);
  });

  test('parses marketplace request with nullable own offer', () {
    final request = MarketplaceRequest.fromJson({'id':'ride-1','pickup':{'latitude':33.6844,'longitude':73.0479},'destination':{'latitude':33.5651,'longitude':73.0169},'proposed_fare':{'amount_minor':10000,'currency':'PKR'},'created_at':'2026-09-15T10:00:00Z','expires_at':'2026-09-15T10:10:00Z','response_deadline':'2026-09-15T10:02:00Z','own_offer':null,'pickup_distance_meters':1250});
    expect(request.ownOffer, isNull);
    expect(request.responseDeadline, DateTime.parse('2026-09-15T10:02:00Z'));
  });

  test('parses expired Rider ride request', () {
    final ride = RiderRideSnapshot.fromJson({'service_code':'car','id':'ride-expired','pickup':{'latitude':33.6844,'longitude':73.0479},'destination':{'latitude':33.5651,'longitude':73.0169},'proposed_fare':{'amount_minor':10000,'currency':'PKR'},'status':'expired','created_at':'2026-09-15T10:00:00Z','expires_at':'2026-09-15T10:10:00Z','trip':null});
    expect(ride.status, 'expired');
    expect(ride.trip, isNull);
  });

  test('allows nullable driver location in aggregate snapshot', () {
    final snapshot = RideSnapshot(ride:RiderRideSnapshot.fromJson({'service_code':'car','id':'ride-1','pickup':{'latitude':33.6844,'longitude':73.0479},'destination':{'latitude':33.5651,'longitude':73.0169},'proposed_fare':{'amount_minor':10000,'currency':'PKR'},'status':'requested','created_at':'2026-09-15T10:00:00Z','expires_at':'2026-09-15T10:10:00Z','trip':null}),offers:const [],driverLocation:null);
    expect(snapshot.driverLocation, isNull);
  });

  test('parses driver location response', () {
    final location = DriverLocationSnapshot.fromJson({'ride_request_id':'ride-1','latitude':33.6844,'longitude':73.0479,'updated_at':'2026-09-15T10:03:00Z'});
    expect(location.rideRequestId, 'ride-1');
  });

  test('parses driver trip history item', () {
    final history = DriverTripHistoryItem.fromJson({'operation_context':null,'ride_request_id':'ride-1','pickup':{'latitude':33.6844,'longitude':73.0479},'destination':{'latitude':33.5651,'longitude':73.0169},'status':'completed','assigned_at':'2026-09-15T10:02:00Z','started_at':'2026-09-15T10:05:00Z','completed_at':'2026-09-15T10:25:00Z','cancelled_at':null,'settlement':{'status':'cash_collected','method':'cash','cash_collected_at':'2026-09-15T10:26:00Z'}});
    expect(history.rideRequestId, 'ride-1');
    expect(history.status, 'completed');
    expect(history.settlement.status, 'cash_collected');
  });
}
