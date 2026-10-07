import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../../rider_request/domain/ride_request.dart';
import '../../rider_request/domain/route_preview.dart';

class DriverRouteCancelled implements Exception {
  const DriverRouteCancelled();
}

abstract interface class DriverRouteRepository {
  Future<RoutePreview> preview({
    required String rideRequestId,
    required String status,
    GeoPoint? origin,
  });

  void cancel();
}

class ApiDriverRouteRepository implements DriverRouteRepository {
  ApiDriverRouteRepository(this._dio, this._sessions);

  final Dio _dio;
  final SessionStore _sessions;
  CancelToken? _activeRequest;

  @override
  Future<RoutePreview> preview({
    required String rideRequestId,
    required String status,
    GeoPoint? origin,
  }) async {
    cancel();
    final cancelToken = CancelToken();
    _activeRequest = cancelToken;
    try {
      final token = await _sessions.readValidToken();
      if (token == null) {
        throw const ApiException(
          'authentication_required',
          'Please sign in again.',
          statusCode: 401,
        );
      }
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/driver/trip/route-preview',
        data: {
          'ride_request_id': rideRequestId,
          'status': status,
          if (origin != null)
            'origin': {
              'latitude': origin.latitude,
              'longitude': origin.longitude,
            },
        },
        cancelToken: cancelToken,
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );
      return RoutePreview.fromJson(response.data!);
    } on DioException catch (error) {
      if (CancelToken.isCancel(error)) {
        throw const DriverRouteCancelled();
      }
      throw ApiException.fromDio(error);
    } finally {
      if (identical(_activeRequest, cancelToken)) {
        _activeRequest = null;
      }
    }
  }

  @override
  void cancel() {
    _activeRequest?.cancel('Driver route superseded');
    _activeRequest = null;
  }
}
