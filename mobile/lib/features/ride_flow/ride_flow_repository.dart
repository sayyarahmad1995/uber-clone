import 'package:dio/dio.dart';

import '../../core/network/api_exception.dart';
import '../../core/session/session_store.dart';

typedef RideFlowPayload = Map<String, dynamic>;

abstract interface class RideFlowRepository {
  Future<RideFlowPayload?> loadDriverTrip();
  Future<RideFlowPayload> loadDriverOpportunities();
  Future<RideFlowPayload> loadDriverHistory();
  Future<RideFlowPayload> loadRiderRide(String rideId);
  Future<RideFlowPayload> loadRiderHistory();
  Future<RideFlowPayload> loadRiderOffers(String rideId);
  Future<RideFlowPayload?> loadDriverLocation(String rideId);
  Future<void> acceptFare(String rideId);
  Future<void> proposeFare(String rideId, int amountMinor);
  Future<void> selectOffer(
    String rideId,
    String driverUserId,
    String updatedAt,
  );
  Future<void> rejectOffer(String rideId, String driverUserId);
  Future<void> startTrip(String rideId);
  Future<void> completeTrip(String rideId);
  Future<void> cancelTrip(String rideId);
  Future<void> confirmCashCollected(String rideId);
}

class ApiRideFlowRepository implements RideFlowRepository {
  ApiRideFlowRepository(this.dio, this.sessions);
  final Dio dio;
  final SessionStore sessions;
  Future<RideFlowPayload> _request(
    String path,
    String method,
    RideFlowPayload? data,
  ) async {
    final token = await sessions.readValidToken();
    if (token == null) {
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }
    try {
      final response = await dio.request<RideFlowPayload>(
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

  Future<void> _post(
    String path, {
    RideFlowPayload? data,
    bool put = false,
  }) async => _request(path, put ? 'PUT' : 'POST', data);
  @override
  Future<RideFlowPayload?> loadDriverTrip() async {
    try {
      return await _request('/v1/driver/trip', 'GET', null);
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<RideFlowPayload> loadDriverOpportunities() =>
      _request('/v1/driver/marketplace/ride-requests', 'GET', null);
  @override
  Future<RideFlowPayload> loadDriverHistory() =>
      _request('/v1/driver/trips', 'GET', null);
  @override
  Future<RideFlowPayload> loadRiderRide(String id) =>
      _request('/v1/ride-requests/$id', 'GET', null);
  @override
  Future<RideFlowPayload> loadRiderHistory() =>
      _request('/v1/ride-requests', 'GET', null);
  @override
  Future<RideFlowPayload> loadRiderOffers(String id) =>
      _request('/v1/ride-requests/$id/offers', 'GET', null);
  @override
  Future<RideFlowPayload?> loadDriverLocation(String id) async {
    try {
      return await _request(
        '/v1/ride-requests/$id/driver-location',
        'GET',
        null,
      );
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<void> acceptFare(String id) =>
      _post('/v1/driver/ride-requests/$id/accept');
  @override
  Future<void> proposeFare(String id, int amount) => _post(
    '/v1/driver/ride-requests/$id/offer',
    data: {'amount_minor': amount},
    put: true,
  );
  @override
  Future<void> selectOffer(String id, String driver, String updatedAt) => _post(
    '/v1/ride-requests/$id/offers/$driver/accept',
    data: {'updated_at': updatedAt},
  );
  @override
  Future<void> rejectOffer(String id, String driver) =>
      _post('/v1/ride-requests/$id/offers/$driver/reject');
  @override
  Future<void> startTrip(String id) =>
      _post('/v1/driver/ride-requests/$id/start');
  @override
  Future<void> completeTrip(String id) =>
      _post('/v1/driver/ride-requests/$id/complete');
  @override
  Future<void> cancelTrip(String id) =>
      _post('/v1/driver/ride-requests/$id/cancel');
  @override
  Future<void> confirmCashCollected(String id) =>
      _post('/v1/driver/ride-requests/$id/cash-collected');
}
