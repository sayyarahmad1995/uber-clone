import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_panels.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';
import 'package:uber_clone/features/rider_request/data/ride_service_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_service.dart';

import 'test_doubles.dart';

void main() {
  const third = RideService(
    code: 'third_car',
    displayName: 'Third category',
    description: 'Test-only service',
    displayOrder: 15,
    presentationToken: 'future-icon',
  );

  test(
    'Rider catalog contract preserves future service codes and tokens',
    () async {
      final adapter = _RideServicesAdapter();
      final repository = ApiRideServiceRepository(
        Dio(BaseOptions(baseUrl: 'http://application.test'))
          ..httpClientAdapter = adapter,
        _SessionStore(),
      );

      final services = await repository.list();
      expect(adapter.authorization, 'Bearer catalog-token');
      expect(adapter.path, '/v1/ride-services');
      expect(services.map((value) => value.code), [
        'economy',
        'third_car',
        'comfort',
      ]);
      expect(services[1].presentationToken, 'future-icon');
      expect(services[1].displayOrder, 15);
    },
  );

  testWidgets('Rider can select a service that was not compiled into Flutter', (
    tester,
  ) async {
    final controller = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          riderRequestControllerProvider.overrideWith((ref) => controller),
          riderServicesProvider.overrideWith(
            (ref) async => [
              const RideService(
                code: 'economy',
                displayName: 'Economy',
                description: 'Standard',
                displayOrder: 10,
                presentationToken: 'car',
              ),
              third,
            ],
          ),
        ],
        child: const MaterialApp(home: Scaffold(body: RiderServicePicker())),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byType(DropdownButtonFormField<String>));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Third category').last);
    await tester.pumpAndSettle();

    expect(controller.serviceCode, 'third_car');
    expect(find.text('Third category'), findsOneWidget);
  });

  testWidgets('A disabled selection is cleared after catalog refresh', (
    tester,
  ) async {
    final controller = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    controller.selectService('third_car');
    final catalog = _MutableCatalog([third]);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          riderRequestControllerProvider.overrideWith((ref) => controller),
          riderServiceRepositoryProvider.overrideWith((ref) => catalog),
        ],
        child: const MaterialApp(home: Scaffold(body: RiderServicePicker())),
      ),
    );
    await tester.pumpAndSettle();
    expect(controller.serviceCode, 'third_car');

    catalog.services = [
      const RideService(
        code: 'economy',
        displayName: 'Economy',
        description: '',
        displayOrder: 10,
        presentationToken: 'car',
      ),
    ];
    final context = tester.element(find.byType(RiderServicePicker));
    ProviderScope.containerOf(context).invalidate(riderServicesProvider);
    await tester.pumpAndSettle();

    expect(controller.serviceCode, isEmpty);
    expect(find.text('Choose an available service'), findsOneWidget);
  });
}

class _MutableCatalog implements RideServiceRepository {
  _MutableCatalog(this.services);

  List<RideService> services;

  @override
  Future<List<RideService>> list() async => List.of(services);
}

class _SessionStore implements SessionStore {
  @override
  Future<String?> readValidToken() async => 'catalog-token';

  @override
  Future<void> save(String token, DateTime expiresAt) async {}

  @override
  Future<void> clear() async {}
}

class _RideServicesAdapter implements HttpClientAdapter {
  String? path;
  String? authorization;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    path = options.path;
    authorization = options.headers['Authorization'] as String?;
    return ResponseBody.fromString(
      jsonEncode({
        'services': [
          {
            'code': 'economy',
            'display_name': 'Economy',
            'description': 'Standard',
            'display_order': 10,
            'presentation_token': 'car',
          },
          {
            'code': 'third_car',
            'display_name': 'Third category',
            'description': 'Test-only service',
            'display_order': 15,
            'presentation_token': 'future-icon',
          },
          {
            'code': 'comfort',
            'display_name': 'Comfort',
            'description': 'Another service',
            'display_order': 20,
            'presentation_token': 'car-front',
          },
        ],
      }),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
