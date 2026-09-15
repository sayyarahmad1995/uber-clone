import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_flow_models.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';

class SemanticRideFlowFake implements RideFlowRepository {
  @override
  Future<List<MarketplaceRequest>> listMarketplaceRequests() async => const [];

  @override
  Future<TripSnapshot?> getCurrentDriverTrip() async => null;

  @override
  Future<List<DriverTripHistoryItem>> listDriverTrips() async => const [];

  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) =>
      throw UnimplementedError();

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

void main() {
  test('ride flow repository exposes semantic operations instead of raw paths', () {
    final repository = SemanticRideFlowFake();
    expect(repository, isA<RideFlowRepository>());
  });
}
