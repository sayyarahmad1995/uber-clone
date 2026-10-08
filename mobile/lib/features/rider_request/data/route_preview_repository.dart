import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../domain/ride_request.dart';
import '../domain/route_preview.dart';

class RoutePreviewCancelled implements Exception {
  const RoutePreviewCancelled();
}

abstract interface class RoutePreviewRepository {
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  });

  void cancel();
}

class ApiRoutePreviewRepository implements RoutePreviewRepository {
  ApiRoutePreviewRepository(this._dio, this._sessions);

  final Dio _dio;
  final SessionStore _sessions;
  CancelToken? _activeRequest;

  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  }) async {
    cancel();
    final cancelToken = CancelToken();
    _activeRequest = cancelToken;

    final token = await _sessions.readValidToken();
    if (token == null) {
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }

    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/ride-previews',
        data: {
          'pickup': {
            'latitude': pickup.latitude,
            'longitude': pickup.longitude,
            if (pickupPlaceId != null && pickupPlaceId.trim().isNotEmpty)
              'place_id': pickupPlaceId.trim(),
          },
          'destination': {
            'latitude': destination.latitude,
            'longitude': destination.longitude,
            if (destinationPlaceId != null &&
                destinationPlaceId.trim().isNotEmpty)
              'place_id': destinationPlaceId.trim(),
          },
          'service_code': serviceCode,
        },
        cancelToken: cancelToken,
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );
      return RoutePreview.fromJson(response.data!);
    } on DioException catch (error) {
      if (CancelToken.isCancel(error)) {
        throw const RoutePreviewCancelled();
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
    _activeRequest?.cancel('route preview superseded');
    _activeRequest = null;
  }
}
