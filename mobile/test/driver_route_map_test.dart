import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/maps/ride_map.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/driver_workspace/presentation/driver_workspace_screen.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_execution.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('Driver route switches from pickup to destination on Start trip', (
    tester,
  ) async {
    final flow = _Flow();
    final routes = _Routes();
    final container = await _pump(tester, flow, routes);
    expect(_map(tester).polylines, hasLength(1));
    expect(_map(tester).polylines.single.points.first.latitude, 38.5);
    expect(routes.requests.single['status'], 'assigned');
    expect(routes.requests.single['origin'], {
      'latitude': 24.86,
      'longitude': 67.01,
    });

    await container.read(driverTripControllerProvider).startTrip('ride-1');
    await tester.pumpAndSettle();
    expect(routes.requests.last['status'], 'in_progress');
    expect(routes.requests.last.containsKey('origin'), isFalse);
    expect(_map(tester).polylines, hasLength(1));
    expect(_map(tester).polylines.single.points.first.latitude, 0);

    await container.read(driverTripControllerProvider).completeTrip('ride-1');
    await tester.pumpAndSettle();
    expect(_map(tester).polylines, isEmpty);
    expect(routes.requests, hasLength(2));
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('restored in-progress trip loads its route without pickup GPS', (
    tester,
  ) async {
    final flow = _Flow()..currentTrip = _trip('in_progress');
    final routes = _Routes();
    await _pump(tester, flow, routes);
    expect(routes.requests.single['status'], 'in_progress');
    expect(routes.requests.single.containsKey('origin'), isFalse);
    expect(_map(tester).polylines.single.points.first.latitude, 0);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('late pickup route cannot overwrite the started-trip route', (
    tester,
  ) async {
    final flow = _Flow();
    final routes = _Routes()..firstReply = Completer<ResponseBody>();
    final container = await _pump(tester, flow, routes);
    expect(_map(tester).polylines, isEmpty);
    await container.read(driverTripControllerProvider).startTrip('ride-1');
    await tester.pumpAndSettle();
    expect(_map(tester).polylines.single.points.first.latitude, 0);
    routes.firstReply!.complete(_response('assigned'));
    await tester.pumpAndSettle();
    expect(_map(tester).polylines.single.points.first.latitude, 0);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('route failure exposes retry and never draws a fabricated route', (
    tester,
  ) async {
    final routes = _Routes()..fail = true;
    await _pump(tester, _Flow(), routes);
    expect(_map(tester).polylines, isEmpty);
    expect(find.byKey(const Key('driverRouteRetryButton')), findsOneWidget);
    routes.fail = false;
    await tester.tap(find.byKey(const Key('driverRouteRetryButton')));
    await tester.pumpAndSettle();
    expect(_map(tester).polylines, hasLength(1));
    expect(find.byKey(const Key('driverRouteRetryButton')), findsNothing);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('cancelled trip clears the route and polling does not reprice it', (
    tester,
  ) async {
    final flow = _Flow();
    final routes = _Routes();
    final container = await _pump(tester, flow, routes);
    await container.read(driverTripControllerProvider).refresh();
    await tester.pumpAndSettle();
    expect(routes.requests, hasLength(1));
    await container.read(driverTripControllerProvider).cancelTrip('ride-1');
    await tester.pumpAndSettle();
    expect(_map(tester).polylines, isEmpty);
    expect(routes.requests, hasLength(1));
    await tester.pumpWidget(const SizedBox.shrink());
  });
}

RideMap _map(WidgetTester tester) =>
    tester.widget<RideMap>(find.byType(RideMap));

Future<ProviderContainer> _pump(
  WidgetTester tester,
  _Flow flow,
  _Routes routes,
) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        authRepositoryProvider.overrideWithValue(
          FakeAuthRepository(account: bothCapabilities),
        ),
        capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
        dioProvider.overrideWithValue(
          Dio(BaseOptions(baseUrl: 'http://application.test'))
            ..httpClientAdapter = routes,
        ),
        sessionStoreProvider.overrideWithValue(_Sessions()),
        rideFlowRepositoryProvider.overrideWithValue(flow),
        driverRepositoryProvider.overrideWithValue(
          FakeDriverRepository(profile: driverProfile),
        ),
        driverPresenceServiceProvider.overrideWithValue(
          FakeDriverPresenceService(),
        ),
        deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
      ],
      child: const MaterialApp(
        home: Scaffold(body: DriverWorkspaceScreen(accountID: 'user-1')),
      ),
    ),
  );
  await tester.pumpAndSettle();
  return ProviderScope.containerOf(
    tester.element(find.byType(DriverWorkspaceScreen)),
  );
}

TripSnapshot _trip(String status) => TripSnapshot(
  rideRequestId: 'ride-1',
  pickup: const RideLocation(latitude: 24.87, longitude: 67.02),
  destination: const RideLocation(latitude: 24.90, longitude: 67.05),
  status: status,
  assignedAt: DateTime.utc(2026, 10, 7),
  settlement: const SettlementSnapshot(status: 'unsettled'),
);

class _Flow extends FakeRideFlowRepository {
  _Flow() : super(currentTrip: _trip('assigned'));
  @override
  Future<void> startTrip(String rideRequestId) async {
    currentTrip = _trip('in_progress');
  }
  @override
  Future<void> completeTrip(String rideRequestId) async {
    currentTrip = _trip('completed');
  }
  @override
  Future<void> cancelDriverTrip(String rideRequestId) async {
    currentTrip = _trip('cancelled');
  }
}

class _Sessions implements SessionStore {
  @override
  Future<String?> readValidToken() async => 'route-test';
  @override
  Future<void> save(String token, DateTime expiresAt) async {}
  @override
  Future<void> clear() async {}
}

class _Routes implements HttpClientAdapter {
  final requests = <Map<String, dynamic>>[];
  Completer<ResponseBody>? firstReply;
  bool fail = false;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    expect(options.path, '/v1/driver/trip/route-preview');
    expect(options.method, 'POST');
    expect(options.headers['Authorization'], isNotEmpty);
    final data = Map<String, dynamic>.from(options.data as Map);
    requests.add(data);
    if (firstReply != null && requests.length == 1) {
      return firstReply!.future;
    }
    if (fail) {
      return ResponseBody.fromString(
        '{"error":"route unavailable"}',
        503,
        headers: {Headers.contentTypeHeader: [Headers.jsonContentType]},
      );
    }
    return _response(data['status'] as String);
  }
  @override
  void close({bool force = false}) {}
}

ResponseBody _response(String status) => ResponseBody.fromString(
  jsonEncode({
    'route': {
      'distance_meters': 2000,
      'duration_seconds': 300,
      'encoded_polyline': status == 'assigned'
          ? '_p~iF~ps|U_ulLnnqC'
          : '??_ibE_ibE',
    },
  }),
  200,
  headers: {Headers.contentTypeHeader: [Headers.jsonContentType]},
);
