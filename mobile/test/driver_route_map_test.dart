import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/maps/ride_map.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/driver_workspace/presentation/driver_workspace_screen.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_execution.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart'
    show GeoPoint;

import 'test_doubles.dart';

void main() {
  const navigationChannel = MethodChannel('higo/navigation');
  setUp(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(navigationChannel, (_) async => true);
  });
  tearDown(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(navigationChannel, null);
  });

  testWidgets('route estimates and navigation follow the current trip stage', (
    tester,
  ) async {
    final launched = <MethodCall>[];
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(navigationChannel, (call) async {
          launched.add(call);
          return true;
        });
    final flow = _Flow();
    final container = await _pump(tester, flow, _Routes());
    expect(
      find.text('Estimated route to pickup: 2.0 km, 5 min'),
      findsOneWidget,
    );
    await _tapNavigation(tester, 'Navigate to pickup');
    await tester.pumpAndSettle();
    expect(launched.single.method, 'openGoogleMaps');
    expect(launched.single.arguments, {'latitude': 24.87, 'longitude': 67.02});

    await container.read(driverTripControllerProvider).startTrip('ride-1');
    await tester.pumpAndSettle();
    expect(find.text('Navigate to pickup'), findsNothing);
    expect(
      find.text('Estimated full trip route: 2.0 km, 10 min'),
      findsOneWidget,
    );
    await _tapNavigation(tester, 'Navigate to destination');
    await tester.pumpAndSettle();
    expect(launched.last.arguments, {'latitude': 24.90, 'longitude': 67.05});
    await container.read(driverTripControllerProvider).completeTrip('ride-1');
    await tester.pumpAndSettle();
    expect(find.text('Navigate to destination'), findsNothing);
    expect(find.textContaining('Estimated '), findsNothing);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('navigation failure is recoverable without changing the trip', (
    tester,
  ) async {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(navigationChannel, (_) async => false);
    final container = await _pump(tester, _Flow(), _Routes());
    await _tapNavigation(tester, 'Navigate to pickup');
    await tester.pumpAndSettle();
    expect(find.text('Unable to open Google Maps. Try again.'), findsOneWidget);
    expect(
      container.read(driverTripControllerProvider).trip?.status,
      'assigned',
    );
    expect(find.text('Navigate to pickup'), findsOneWidget);
    expect(_map(tester).polylines, hasLength(1));
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('navigation remains usable when the HiGO route is unavailable', (
    tester,
  ) async {
    await _pump(tester, _Flow(), _Routes()..fail = true);
    expect(find.textContaining('Estimated '), findsNothing);
    expect(find.text('Navigate to pickup'), findsOneWidget);
    await _tapNavigation(tester, 'Navigate to pickup');
    await tester.pumpAndSettle();
    expect(find.text('Unable to open Google Maps. Try again.'), findsNothing);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets(
    'late navigation failure cannot leak into a different trip stage',
    (tester) async {
      final reply = Completer<bool>();
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(navigationChannel, (_) => reply.future);
      final container = await _pump(tester, _Flow(), _Routes());
      await _tapNavigation(tester, 'Navigate to pickup');
      await tester.pump();
      await container.read(driverTripControllerProvider).startTrip('ride-1');
      await tester.pumpAndSettle();
      reply.complete(false);
      await tester.pumpAndSettle();
      expect(find.text('Unable to open Google Maps. Try again.'), findsNothing);
      expect(find.text('Navigate to destination'), findsOneWidget);
      await container.read(driverTripControllerProvider).cancelTrip('ride-1');
      await tester.pumpAndSettle();
      expect(find.text('Navigate to destination'), findsNothing);
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  testWidgets('navigation does not launch twice while a handoff is pending', (
    tester,
  ) async {
    final reply = Completer<bool>();
    var calls = 0;
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(navigationChannel, (_) {
          calls++;
          return reply.future;
        });
    await _pump(tester, _Flow(), _Routes());
    await _tapNavigation(tester, 'Navigate to pickup');
    await tester.pump();
    await _tapNavigation(tester, 'Navigate to pickup');
    await tester.pump();
    expect(calls, 1);
    reply.complete(true);
    await tester.pumpAndSettle();
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('small expanded panel keeps navigation clear of route errors', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(360, 500);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    var launches = 0;
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(navigationChannel, (_) async {
          launches++;
          return true;
        });
    await _pump(tester, _Flow(), _Routes()..fail = true);
    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -400),
    );
    await tester.pumpAndSettle();
    final navigation = tester.getRect(
      find.byKey(const Key('driverNavigationButton')),
    );
    final status = tester.getRect(find.byType(DashboardStatusCard));
    expect(navigation.overlaps(status), isFalse);
    final panel = tester.getRect(find.byKey(const Key('dashboardPanel')));
    expect(panel.contains(navigation.topLeft), isTrue);
    expect(panel.contains(navigation.bottomRight), isTrue);
    await tester.tap(find.text('Navigate to pickup'));
    await tester.pumpAndSettle();
    expect(launches, 1);
    expect(tester.takeException(), isNull);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets(
    'Driver route switches from pickup to destination on Start trip',
    (tester) async {
      final flow = _Flow();
      final routes = _Routes();
      final container = await _pump(tester, flow, routes);
      expect(_map(tester).polylines, hasLength(1));
      expect(_map(tester).polylines.single.points.first.latitude, 38.5);
      expect(routes.received.single.path, '/v1/driver/trip/route-preview');
      expect(routes.received.single.method, 'POST');
      expect(
        routes.received.single.headers['Authorization'],
        ['Bearer', 'route-test'].join(' '),
      );
      expect(routes.requests.single['status'], 'assigned');
      expect(routes.requests.single['origin'], {
        'latitude': 24.86,
        'longitude': 67.01,
      });

      final tripController = container.read(driverTripControllerProvider);
      await tripController.startTrip('ride-1');
      expect(tripController.error, isNull);
      expect(tripController.trip?.status, 'in_progress');
      await tester.pumpAndSettle();
      expect(
        container.read(driverRouteControllerProvider).state.status,
        'in_progress',
      );
      expect(container.read(driverRouteControllerProvider).state.error, isNull);
      expect(
        container.read(driverRouteControllerProvider).state.loading,
        isFalse,
      );
      expect(routes.requests.last['status'], 'in_progress');
      expect(
        routes.received.last.headers['Authorization'],
        ['Bearer', 'route-test'].join(' '),
      );
      expect(routes.requests.last.containsKey('origin'), isFalse);
      expect(_map(tester).polylines, hasLength(1));
      expect(_map(tester).polylines.single.points.first.latitude, 0);

      await container.read(driverTripControllerProvider).completeTrip('ride-1');
      await tester.pumpAndSettle();
      expect(_map(tester).polylines, isEmpty);
      expect(routes.requests, hasLength(2));
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

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

  testWidgets(
    'route failure exposes retry and never draws a fabricated route',
    (tester) async {
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
    },
  );

  testWidgets('late GPS cannot start a pickup route after Start trip', (
    tester,
  ) async {
    final location = _DelayedLocation();
    final routes = _Routes();
    final container = await _pump(tester, _Flow(), routes, location: location);
    expect(routes.requests, isEmpty);
    await container.read(driverTripControllerProvider).startTrip('ride-1');
    await tester.pumpAndSettle();
    expect(_map(tester).polylines.single.points.first.latitude, 0);
    location.reply.complete(const GeoPoint(latitude: 24.86, longitude: 67.01));
    await tester.pumpAndSettle();
    expect(routes.requests, hasLength(1));
    expect(routes.requests.single['status'], 'in_progress');
    expect(_map(tester).polylines.single.points.first.latitude, 0);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('GPS failure leaves no pickup line and offers Retry', (
    tester,
  ) async {
    final routes = _Routes();
    await _pump(tester, _Flow(), routes, location: _FailedLocation());
    expect(routes.requests, isEmpty);
    expect(_map(tester).polylines, isEmpty);
    expect(find.text('Location permission is required.'), findsOneWidget);
    expect(find.byKey(const Key('driverRouteRetryButton')), findsOneWidget);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets(
    'cancelled trip clears the route without refetching unchanged stages',
    (tester) async {
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
    },
  );
}

Future<void> _tapNavigation(WidgetTester tester, String label) async {
  await tester.drag(
    find.byKey(const Key('dashboardPanelDragHandle')),
    const Offset(0, -400),
  );
  await tester.pumpAndSettle();
  await tester.ensureVisible(find.text(label));
  await tester.tap(find.text(label));
}

RideMap _map(WidgetTester tester) =>
    tester.widget<RideMap>(find.byType(RideMap));

Future<ProviderContainer> _pump(
  WidgetTester tester,
  _Flow flow,
  _Routes routes, {
  DeviceLocation? location,
}) async {
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
        deviceLocationProvider.overrideWithValue(
          location ?? const FakeDeviceLocation(),
        ),
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
  final received = <RequestOptions>[];
  Completer<ResponseBody>? firstReply;
  bool fail = false;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final data = Map<String, dynamic>.from(options.data as Map);
    received.add(options);
    requests.add(data);
    if (firstReply != null && requests.length == 1) {
      return firstReply!.future;
    }
    if (fail) {
      return ResponseBody.fromString(
        '{"error":"route unavailable"}',
        503,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
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
      'duration_seconds': status == 'assigned' ? 300 : 600,
      'encoded_polyline': status == 'assigned'
          ? '_p~iF~ps|U_ulLnnqC'
          : '??_ibE_ibE',
    },
  }),
  200,
  headers: {
    Headers.contentTypeHeader: [Headers.jsonContentType],
  },
);

class _DelayedLocation implements DeviceLocation {
  final reply = Completer<GeoPoint>();
  @override
  Future<GeoPoint> current() => reply.future;
}

class _FailedLocation implements DeviceLocation {
  @override
  Future<GeoPoint> current() async =>
      throw const LocationUnavailable('Location permission is required.');
}
