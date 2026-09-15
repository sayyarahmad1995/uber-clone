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
  Future<List<RiderRideSnapshot>> listRiderRides();
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId);
  Future<List<RiderOfferComparison>> listRiderOffers(String rideRequestId);
  Future<DriverLocationSnapshot?> getDriverLocation(String rideRequestId);
  Future<void> submitOffer(String rideRequestId, int amountMinor);
  Future<void> acceptProposedFare(String rideRequestId);
  Future<void> selectOffer(
    String rideRequestId,
    String driverUserId,
    DateTime updatedAt,
  );
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
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }
    try {
      final response = await dio.request<Json>(
        path,
        data: data,
        options: Options(
          method: method,
          headers: {'Authorization': 'Bearer $token'},
        ),
      );
      return response.data ?? <String, dynamic>{};
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  List<T> _list<T>(Json json, String key, T Function(Json) parse) =>
      (json[key] as List? ?? const [])
          .map((value) => parse(Map<String, dynamic>.from(value as Map)))
          .toList();

  @override
  Future<List<MarketplaceRequest>> listMarketplaceRequests() async => _list(
    await _request('/v1/driver/marketplace/ride-requests', 'GET', null),
    'ride_requests',
    MarketplaceRequest.fromJson,
  );

  @override
  Future<TripSnapshot?> getCurrentDriverTrip() async {
    try {
      final json = await _request('/v1/driver/trip', 'GET', null);
      return TripSnapshot.fromJson(json);
    } on ApiException catch (error) {
      if (error.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<List<DriverTripHistoryItem>> listDriverTrips() async => _list(
    await _request('/v1/driver/trips', 'GET', null),
    'trips',
    DriverTripHistoryItem.fromJson,
  );

  @override
  Future<List<RiderRideSnapshot>> listRiderRides() async => _list(
    await _request('/v1/ride-requests', 'GET', null),
    'ride_requests',
    RiderRideSnapshot.fromJson,
  );

  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) async =>
      RiderRideSnapshot.fromJson(
        await _request('/v1/ride-requests/$rideRequestId', 'GET', null),
      );

  @override
  Future<List<RiderOfferComparison>> listRiderOffers(
    String rideRequestId,
  ) async => _list(
    await _request('/v1/ride-requests/$rideRequestId/offers', 'GET', null),
    'offers',
    RiderOfferComparison.fromJson,
  );

  @override
  Future<DriverLocationSnapshot?> getDriverLocation(String rideRequestId) async {
    try {
      return DriverLocationSnapshot.fromJson(
        await _request(
          '/v1/ride-requests/$rideRequestId/driver-location',
          'GET',
          null,
        ),
      );
    } on ApiException catch (error) {
      if (error.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<void> submitOffer(String rideRequestId, int amountMinor) async {
    await _request(
      '/v1/driver/ride-requests/$rideRequestId/offer',
      'PUT',
      {'amount_minor': amountMinor},
    );
  }

  @override
  Future<void> acceptProposedFare(String rideRequestId) async {
    await _request(
      '/v1/driver/ride-requests/$rideRequestId/accept-proposed-fare',
      'POST',
      null,
    );
  }

  @override
  Future<void> selectOffer(
    String rideRequestId,
    String driverUserId,
    DateTime updatedAt,
  ) async {
    await _request(
      '/v1/ride-requests/$rideRequestId/offers/$driverUserId/accept',
      'POST',
      {'updated_at': updatedAt.toUtc().toIso8601String()},
    );
  }

  @override
  Future<void> startTrip(String rideRequestId) async {
    await _request(
      '/v1/driver/ride-requests/$rideRequestId/start',
      'POST',
      null,
    );
  }

  @override
  Future<void> completeTrip(String rideRequestId) async {
    await _request(
      '/v1/driver/ride-requests/$rideRequestId/complete',
      'POST',
      null,
    );
  }

  @override
  Future<void> confirmCashCollected(String rideRequestId) async {
    await _request(
      '/v1/driver/ride-requests/$rideRequestId/cash-collected',
      'POST',
      null,
    );
  }

  @override
  Future<void> cancelDriverTrip(String rideRequestId) async {
    await _request(
      '/v1/driver/ride-requests/$rideRequestId/cancel',
      'POST',
      null,
    );
  }
}
