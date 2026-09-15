import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/driver_onboarding/domain/driver_onboarding_application.dart';
import 'package:uber_clone/features/driver_onboarding/domain/driver_service.dart';
import 'package:uber_clone/features/driver_onboarding/domain/driver_vehicle.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_offer.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_snapshot.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';

// Existing dashboard tests isolate networking; flow behavior has dedicated tests.
class FakeRideFlowRepository implements RideFlowRepository {
  @override
  Future<List<MarketplaceRequest>> listMarketplaceRequests() async => const [];

  @override
  Future<TripSnapshot?> getCurrentDriverTrip() async => null;

  @override
  Future<List<DriverTripHistoryItem>> listDriverTrips() async => const [];

  @override
  Future<List<RiderRideSnapshot>> listRiderRides() async => const [];

  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) {
    throw const ApiException('not_found', 'No ride', statusCode: 404);
  }

  @override
  Future<List<RiderOfferComparison>> listRiderOffers(String rideRequestId) async =>
      const [];

  @override
  Future<DriverLocationSnapshot?> getDriverLocation(String rideRequestId) async =>
      null;

  @override
  Future<void> submitOffer(String rideRequestId, int amountMinor) async {}

  @override
  Future<void> acceptProposedFare(String rideRequestId) async {}

  @override
  Future<void> selectOffer(
    String rideRequestId,
    String driverUserId,
    DateTime updatedAt,
  ) async {}

  @override
  Future<void> startTrip(String rideRequestId) async {}

  @override
  Future<void> completeTrip(String rideRequestId) async {}

  @override
  Future<void> confirmCashCollected(String rideRequestId) async {}

  @override
  Future<void> cancelDriverTrip(String rideRequestId) async {}
}
