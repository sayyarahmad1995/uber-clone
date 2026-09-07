import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/driver_workspace/data/driver_onboarding_repository.dart';

import 'ride_request_repository_test.dart' show StaticSessionStore;
import 'test_doubles.dart';

void main() {
  late DriverOnboardingAdapter adapter;
  late ApiDriverOnboardingRepository repository;

  setUp(() {
    adapter = DriverOnboardingAdapter();
    final dio = Dio(BaseOptions(baseUrl: 'http://application.test'))
      ..httpClientAdapter = adapter;
    repository = ApiDriverOnboardingRepository(dio, StaticSessionStore());
  });

  test(
    'loads service catalog and treats onboarding 404 as no application',
    () async {
      final services = await repository.listServices();
      expect(services.map((service) => service.code), ['economy', 'comfort']);
      expect(adapter.request!.method, 'GET');
      expect(adapter.request!.path, '/v1/driver/services');

      adapter.hasApplication = false;
      expect(await repository.getLatest(), isNull);
      expect(adapter.request!.path, '/v1/driver/onboarding');
    },
  );

  test('precheck and submit use POST application contracts', () async {
    final precheck = await repository.precheck(
      displayName: ' Test Driver ',
      serviceCode: 'comfort',
      vehicle: driverVehicle,
    );
    expect(precheck.eligible, isTrue);
    expect(adapter.request!.method, 'POST');
    expect(adapter.request!.path, '/v1/driver/onboarding/precheck');
    expect(adapter.request!.data, {
      'display_name': 'Test Driver',
      'service_code': 'comfort',
      'vehicle': {
        'make': 'Toyota',
        'model': 'Corolla',
        'model_year': 2024,
        'color': 'White',
        'license_plate': 'ABC-123',
      },
    });

    final application = await repository.submit(
      displayName: ' Test Driver ',
      serviceCode: 'comfort',
      vehicle: driverVehicle,
    );
    expect(adapter.request!.method, 'POST');
    expect(adapter.request!.path, '/v1/driver/onboarding');
    expect(application.status, 'pending');
    expect(application.service.code, 'comfort');
    expect(application.vehicle, driverVehicle);
  });
}

class DriverOnboardingAdapter implements HttpClientAdapter {
  RequestOptions? request;
  bool hasApplication = true;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    request = options;
    Object data;
    var status = 200;
    if (options.path == '/v1/driver/services') {
      data = {
        'services': [
          {
            'code': 'economy',
            'display_name': 'Economy',
            'description': 'Vehicle eligibility is confirmed during review.',
            'minimum_model_year': null,
            'implied_service_code': null,
          },
          {
            'code': 'comfort',
            'display_name': 'Comfort',
            'description': 'Vehicle condition and service eligibility are confirmed during review.',
            'minimum_model_year': null,
            'implied_service_code': 'economy',
          },
        ],
      };
    } else if (options.path == '/v1/driver/onboarding/precheck') {
      data = {
        'eligible': true,
        'reasons': <String>[],
        'service': {
          'code': 'comfort',
          'display_name': 'Comfort',
          'description': 'Vehicle condition and service eligibility are confirmed during review.',
          'minimum_model_year': null,
          'implied_service_code': 'economy',
        },
      };
    } else if (options.method == 'GET' && !hasApplication) {
      status = 404;
      data = {'error': 'driver onboarding application not found'};
    } else {
      data = {
        'id': 'application-1',
        'display_name': 'Test Driver',
        'status': 'pending',
        'service': {
          'code': 'comfort',
          'display_name': 'Comfort',
          'description': 'Vehicle condition and service eligibility are confirmed during review.',
          'minimum_model_year': null,
          'implied_service_code': 'economy',
        },
        'vehicle': {
          'make': 'Toyota',
          'model': 'Corolla',
          'model_year': 2024,
          'color': 'White',
          'license_plate': 'ABC-123',
        },
        'rejection_reason': '',
        'submitted_at': '2026-09-07T00:00:00Z',
        'decided_at': null,
      };
    }
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
