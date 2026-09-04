import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../domain/ride_request.dart';

abstract interface class RideRequestRepository {
  Future<List<RideRequest>> list();
  Future<RideRequest> create({
    required GeoPoint pickup,
    required GeoPoint destination,
    required Money proposedFare,
  });
  Future<RideRequest> get(String id);
  Future<void> cancel(String id);
}

class ApiRideRequestRepository implements RideRequestRepository {
  ApiRideRequestRepository(this._dio, this._sessions);
  final Dio _dio;
  final SessionStore _sessions;

  Future<Options> _options() async {
    final token = await _sessions.readValidToken();
    if (token == null) {
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }
    return Options(headers: {'Authorization': 'Bearer $token'});
  }

  @override
  Future<List<RideRequest>> list() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/ride-requests',
        options: await _options(),
      );
      return RideRequestList.fromJson(response.data!).rideRequests;
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<RideRequest> create({
    required GeoPoint pickup,
    required GeoPoint destination,
    required Money proposedFare,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/ride-requests',
        options: await _options(),
        data: {
          'pickup': pickup.toJson(),
          'destination': destination.toJson(),
          'proposed_fare': proposedFare.toJson(),
        },
      );
      return RideRequest.fromJson(response.data!);
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<RideRequest> get(String id) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/ride-requests/$id',
        options: await _options(),
      );
      return RideRequest.fromJson(response.data!);
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<void> cancel(String id) async {
    try {
      await _dio.post<void>(
        '/v1/ride-requests/$id/cancel',
        options: await _options(),
      );
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }
}
