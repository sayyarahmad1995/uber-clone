import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../domain/place_search.dart';
import '../domain/ride_request.dart';

class PlaceSearchCancelled implements Exception {
  const PlaceSearchCancelled();
}

abstract interface class PlaceSearchRepository {
  Future<List<PlaceSuggestion>> autocomplete({
    required String input,
    required String sessionToken,
    GeoPoint? bias,
  });

  Future<PlaceSelection> details({
    required String placeId,
    required String sessionToken,
  });

  Future<PlaceSelection?> reverseGeocode(GeoPoint point);

  void cancelSession(String sessionToken);
}

class ApiPlaceSearchRepository implements PlaceSearchRepository {
  ApiPlaceSearchRepository(this._dio, this._sessions);

  final Dio _dio;
  final SessionStore _sessions;
  final Map<String, CancelToken> _autocompleteTokens = {};

  @override
  Future<List<PlaceSuggestion>> autocomplete({
    required String input,
    required String sessionToken,
    GeoPoint? bias,
  }) async {
    final previous = _autocompleteTokens.remove(sessionToken);
    previous?.cancel('superseded autocomplete request');

    final cancelToken = CancelToken();
    _autocompleteTokens[sessionToken] = cancelToken;
    try {
      final response = await _request<Map<String, dynamic>>(
        '/v1/places/autocomplete',
        method: 'POST',
        data: {
          'input': input,
          'session_token': sessionToken,
          if (bias != null)
            'bias': {'latitude': bias.latitude, 'longitude': bias.longitude},
        },
        cancelToken: cancelToken,
      );
      final raw = response['suggestions'] as List? ?? const [];
      return raw
          .map((item) => PlaceSuggestion.fromJson(item as Map<String, dynamic>))
          .toList(growable: false);
    } finally {
      if (identical(_autocompleteTokens[sessionToken], cancelToken)) {
        _autocompleteTokens.remove(sessionToken);
      }
    }
  }

  @override
  Future<PlaceSelection> details({
    required String placeId,
    required String sessionToken,
  }) async {
    final data = await _request<Map<String, dynamic>>(
      '/v1/places/${Uri.encodeComponent(placeId)}',
      queryParameters: {'session_token': sessionToken},
    );
    return PlaceSelection.fromJson(data);
  }

  @override
  Future<PlaceSelection?> reverseGeocode(GeoPoint point) async {
    try {
      final data = await _request<Map<String, dynamic>>(
        '/v1/places/reverse-geocode',
        method: 'POST',
        data: {'latitude': point.latitude, 'longitude': point.longitude},
      );
      return PlaceSelection.fromJson(data);
    } on ApiException catch (error) {
      if (error.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  void cancelSession(String sessionToken) {
    _autocompleteTokens.remove(sessionToken)?.cancel('session ended');
  }

  Future<T> _request<T>(
    String path, {
    String method = 'GET',
    Object? data,
    Map<String, dynamic>? queryParameters,
    CancelToken? cancelToken,
  }) async {
    final token = await _sessions.readValidToken();
    if (token == null) {
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }
    try {
      final response = await _dio.request<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        cancelToken: cancelToken,
        options: Options(
          method: method,
          headers: {'Authorization': 'Bearer $token'},
        ),
      );
      return response.data as T;
    } on DioException catch (error) {
      if (CancelToken.isCancel(error)) {
        throw const PlaceSearchCancelled();
      }
      throw ApiException.fromDio(error);
    }
  }
}
