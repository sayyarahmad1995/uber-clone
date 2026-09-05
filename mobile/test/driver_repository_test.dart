import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/driver_workspace/data/driver_repository.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';
import 'package:uber_clone/core/models/account.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'ride_request_repository_test.dart' show StaticSessionStore;
import 'test_doubles.dart';

void main() {
  late DriverAdapter adapter;
  late ApiDriverRepository repo;
  late Dio dio;
  setUp(() {
    adapter = DriverAdapter();
    dio = Dio(BaseOptions(baseUrl: 'http://application.test'))
      ..httpClientAdapter = adapter;
    repo = ApiDriverRepository(dio, StaticSessionStore());
  });
  test('Driver contracts use authenticated application endpoints', () async {
    await repo.get();
    expect(adapter.request!.method, 'GET');
    expect(adapter.request!.path, '/v1/driver');
    await repo.onboard(' Test Driver ', driverVehicle);
    expect(adapter.request!.method, 'PUT');
    expect(adapter.request!.data, {
      'display_name': 'Test Driver',
      'vehicle': {
        'make': 'Toyota',
        'model': 'Corolla',
        'model_year': 2024,
        'color': 'White',
        'license_plate': 'ABC-123',
      },
    });
    await repo.setOnline(true);
    expect(adapter.request!.path, '/v1/driver/availability');
    expect(adapter.request!.data, {'is_online': true});
    final location = await repo.publishLocation(
      const GeoPoint(latitude: 24, longitude: 67),
    );
    expect(adapter.request!.path, '/v1/driver/location');
    expect(adapter.request!.data, {'latitude': 24.0, 'longitude': 67.0});
    expect(location.updatedAt, DateTime.utc(2026, 9, 5));
    expect(adapter.request!.headers['Authorization'], 'Bearer session-token');
  });
  test('only profile 404 means onboarding required', () async {
    adapter.status = 404;
    expect(await repo.get(), isNull);
    for (final status in [401, 403, 500]) {
      adapter.status = status;
      await expectLater(repo.get(), throwsA(isA<ApiException>()));
    }
  });
  test('legacy presentation nulls remain readable', () async {
    adapter.legacy = true;
    final profile = await repo.get();
    expect(profile!.displayName, isNull);
    expect(profile.vehicle.modelYear, isNull);
  });
  test('enabling Driver capability returns server account', () async {
    final account = await ApiAuthRepository(
      dio,
      StaticSessionStore(),
    ).enableDriver();
    expect(adapter.request!.method, 'PUT');
    expect(adapter.request!.path, '/v1/me/capabilities/driver');
    expect(account.capabilities, contains(Capability.driver));
  });
}

class DriverAdapter implements HttpClientAdapter {
  RequestOptions? request;
  int status = 200;
  bool legacy = false;
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    request = options;
    final Object data = options.path.endsWith('/capabilities/driver')
        ? {
            'id': 'user-1',
            'capabilities': ['rider', 'driver'],
          }
        : options.path.endsWith('/location')
        ? {
            'latitude': 24.0,
            'longitude': 67.0,
            'updated_at': '2026-09-05T00:00:00Z',
          }
        : {
            'user_id': 'user-1',
            'display_name': legacy ? null : 'Test Driver',
            'status': 'active',
            'is_online': false,
            'vehicle': {
              'make': 'Toyota',
              'model': 'Corolla',
              'model_year': legacy ? null : 2024,
              'color': 'White',
              'license_plate': 'ABC-123',
            },
          };
    return ResponseBody.fromString(
      jsonEncode(data),
      status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
