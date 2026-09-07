import 'package:dio/dio.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../domain/driver_onboarding.dart';
import '../domain/driver_profile.dart';

abstract interface class DriverOnboardingRepository {
  Future<List<DriverServiceOption>> listServices();
  Future<DriverOnboardingApplication?> getLatest();
  Future<DriverOnboardingPrecheck> precheck({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  });
  Future<DriverOnboardingApplication> submit({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  });
}

class ApiDriverOnboardingRepository implements DriverOnboardingRepository {
  ApiDriverOnboardingRepository(this._dio, this._sessions);

  final Dio _dio;
  final SessionStore _sessions;

  Future<Map<String, dynamic>> _request(
    String path, {
    String method = 'GET',
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
          method: method,
          headers: {'Authorization': 'Bearer $token'},
        ),
      );
      return response.data!;
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  Map<String, dynamic> _payload({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) => {
    'display_name': displayName.trim(),
    'service_code': serviceCode,
    'vehicle': vehicle.toJson(),
  };

  @override
  Future<List<DriverServiceOption>> listServices() async {
    final body = await _request('/v1/driver/services');
    return (body['services'] as List<dynamic>)
        .map(
          (value) =>
              DriverServiceOption.fromJson(value as Map<String, dynamic>),
        )
        .toList(growable: false);
  }

  @override
  Future<DriverOnboardingApplication?> getLatest() async {
    try {
      return DriverOnboardingApplication.fromJson(
        await _request('/v1/driver/onboarding'),
      );
    } on ApiException catch (error) {
      if (error.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<DriverOnboardingPrecheck> precheck({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async => DriverOnboardingPrecheck.fromJson(
    await _request(
      '/v1/driver/onboarding/precheck',
      method: 'POST',
      data: _payload(
        displayName: displayName,
        serviceCode: serviceCode,
        vehicle: vehicle,
      ),
    ),
  );

  @override
  Future<DriverOnboardingApplication> submit({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async => DriverOnboardingApplication.fromJson(
    await _request(
      '/v1/driver/onboarding',
      method: 'POST',
      data: _payload(
        displayName: displayName,
        serviceCode: serviceCode,
        vehicle: vehicle,
      ),
    ),
  );
}
