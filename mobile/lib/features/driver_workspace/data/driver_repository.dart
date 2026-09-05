import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../../rider_request/domain/ride_request.dart';
import '../domain/driver_profile.dart';

abstract interface class DriverRepository {
  Future<DriverProfile?> get();
  Future<DriverProfile> onboard(String displayName, DriverVehicle vehicle);
  Future<DriverProfile> setOnline(bool online);
  Future<PublishedDriverLocation> publishLocation(GeoPoint point);
}

class ApiDriverRepository implements DriverRepository {
  ApiDriverRepository(this._dio, this._sessions);
  final Dio _dio;
  final SessionStore _sessions;

  Future<Map<String, dynamic>> _request(
    String path, {
    Map<String, dynamic>? data,
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
      final response = await _dio.request<Map<String, dynamic>>(
        path,
        data: data,
        options: Options(
          method: data == null ? 'GET' : 'PUT',
          headers: {'Authorization': 'Bearer $token'},
        ),
      );
      return response.data!;
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<DriverProfile?> get() async {
    try {
      return DriverProfile.fromJson(await _request('/v1/driver'));
    } on ApiException catch (error) {
      if (error.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<DriverProfile> onboard(
    String displayName,
    DriverVehicle vehicle,
  ) async => DriverProfile.fromJson(
    await _request(
      '/v1/driver',
      data: {'display_name': displayName.trim(), 'vehicle': vehicle.toJson()},
    ),
  );

  @override
  Future<DriverProfile> setOnline(bool online) async => DriverProfile.fromJson(
    await _request('/v1/driver/availability', data: {'is_online': online}),
  );

  @override
  Future<PublishedDriverLocation> publishLocation(GeoPoint point) async =>
      PublishedDriverLocation.fromJson(
        await _request('/v1/driver/location', data: point.toJson()),
      );
}
