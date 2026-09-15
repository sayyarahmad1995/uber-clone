import 'package:dio/dio.dart';
import '../../core/network/api_exception.dart';
import '../../core/session/session_store.dart';
import 'domain/marketplace_request.dart';
import 'domain/ride_offer.dart';
import 'domain/ride_snapshot.dart';
import 'domain/trip.dart';

typedef Json = Map<String, dynamic>;

abstract interface class RideFlowRepository {
  Future<List<MarketplaceRequest>> listMarketplaceRequests();
  Future<TripSnapshot?> getCurrentDriverTrip();
  Future<List<DriverTripHistoryItem>> listDriverTrips();
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId);
  Future<List<RiderOfferComparison>> listRiderOffers(String rideRequestId);
  Future<DriverLocationSnapshot?> getDriverLocation(String rideRequestId);
  Future<void> submitOffer(String rideRequestId, int amountMinor);
  Future<void> acceptProposedFare(String rideRequestId);
  Future<void> selectOffer(String rideRequestId, String driverUserId, DateTime updatedAt);
  Future<void> startTrip(String rideRequestId);
  Future<void> completeTrip(String rideRequestId);
  Future<void> confirmCashCollected(String rideRequestId);
  Future<void> cancelDriverTrip(String rideRequestId);
}

class ApiRideFlowRepository implements RideFlowRepository {
  ApiRideFlowRepository(this.dio, this.sessions);
  final Dio dio;
  final SessionStore sessions;

  Future<Json> _request(String path, String method, Json? data) async {
    final token = await sessions.readValidToken();
    if (token == null) {
      throw const ApiException('authentication_required', 'Please sign in again.', statusCode: 401);
    }
    try {
      final response = await dio.request<Json>(
        path,
        data: data,
        options: Options(method: method, headers: {'Authorization': 'Bearer $token'}),
      );
      return response.data ?? <String, dynamic>{};
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<List<MarketplaceRequest>> listMarketplaceRequests() => throw UnimplementedError();
  @override
  Future<TripSnapshot?> getCurrentDriverTrip() => throw UnimplementedError();
  @override
  Future<List<DriverTripHistoryItem>> listDriverTrips() => throw UnimplementedError();
  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) => throw UnimplementedError();
  @override
  Future<List<RiderOfferComparison>> listRiderOffers(String rideRequestId) => throw UnimplementedError();
  @override
  Future<DriverLocationSnapshot?> getDriverLocation(String rideRequestId) => throw UnimplementedError();
  @override
  Future<void> submitOffer(String rideRequestId, int amountMinor) => throw UnimplementedError();
  @override
  Future<void> acceptProposedFare(String rideRequestId) => throw UnimplementedError();
  @override
  Future<void> selectOffer(String rideRequestId, String driverUserId, DateTime updatedAt) => throw UnimplementedError();
  @override
  Future<void> startTrip(String rideRequestId) => throw UnimplementedError();
  @override
  Future<void> completeTrip(String rideRequestId) => throw UnimplementedError();
  @override
  Future<void> confirmCashCollected(String rideRequestId) => throw UnimplementedError();
  @override
  Future<void> cancelDriverTrip(String rideRequestId) => throw UnimplementedError();
}
