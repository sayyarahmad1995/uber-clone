import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../domain/ride_service.dart';

abstract interface class RideServiceRepository {
  Future<List<RideService>> list();
}

class ApiRideServiceRepository implements RideServiceRepository {
  ApiRideServiceRepository(this._dio, this._sessions);

  final Dio _dio;
  final SessionStore _sessions;

  @override
  Future<List<RideService>> list() async {
    final token = await _sessions.readValidToken();
    if (token == null) {
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/ride-services',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );
      final items = response.data!['services'] as List<dynamic>? ?? [];
      return items
          .map(
            (value) =>
                RideService.fromJson(Map<String, dynamic>.from(value as Map)),
          )
          .toList(growable: false);
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }
}
