import 'package:flutter/material.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/rider_request/domain/ride_service.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/maps/ride_map.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_panels.dart';
import 'package:uber_clone/features/rider_request/data/route_preview_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';
import 'package:uber_clone/features/rider_request/domain/route_preview.dart';
import 'package:uber_clone/features/rider_request/presentation/rider_request_screen.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('preselected service can submit after its picker scrolls away', (
    tester,
  ) async {
    final repository = FakeRideRequestRepository();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authRepositoryProvider.overrideWithValue(FakeAuthRepository()),
          capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
          rideRequestRepositoryProvider.overrideWithValue(repository),
          riderServiceRepositoryProvider.overrideWithValue(
            FakeRideServiceRepository(),
          ),
          deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
          routePreviewRepositoryProvider.overrideWithValue(_Routes()),
          rideFlowRepositoryProvider.overrideWithValue(
            FakeRideFlowRepository(),
          ),
        ],
        child: const MaterialApp(home: Scaffold(body: RiderRequestScreen())),
      ),
    );
    await tester.pumpAndSettle();
    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -400),
    );
    await tester.pumpAndSettle();
    expect(find.text('Economy'), findsOneWidget);
    final container = ProviderScope.containerOf(
      tester.element(find.byType(RiderRequestScreen)),
    );
    final controller = container.read(riderRequestControllerProvider);
    controller.setPickup(requestedRide.pickup);
    controller.setDestination(requestedRide.destination);
    await tester.pumpAndSettle();

    expect(
      tester.widget<RideMap>(find.byType(RideMap)).polylines,
      hasLength(1),
    );

    final list = tester.widget<ListView>(find.byType(ListView).first);
    list.controller!.jumpTo(list.controller!.position.maxScrollExtent);
    await tester.pumpAndSettle();
    expect(find.byType(RiderServicePicker), findsNothing);
    expect(controller.serviceCode, 'economy');
    expect(container.read(riderServicesProvider).asData?.value, isNotNull);
    await tester.enterText(find.byKey(const Key('fareField')), '700');
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.byKey(const Key('requestRideButton')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const Key('requestRideButton')).hitTestable());
    await tester.pumpAndSettle();

    expect(find.text('Choose an available ride service.'), findsNothing);
    expect(
      repository.submittedFare,
      const Money(amountMinor: 70000, currency: 'PKR'),
    );
    expect(controller.state.active?.id, 'ride-1');
    final bookedMap = tester.widget<RideMap>(find.byType(RideMap));
    expect(bookedMap.polylines, hasLength(1));
    expect(bookedMap.polylines.single.points, hasLength(2));
    expect(bookedMap.polylines.single.points.first.latitude, 38.5);
    expect(bookedMap.onTap, isNull);
    expect(bookedMap.onPlaceTap, isNull);

    // An unrelated restored request must not display the previous draft route.
    repository.requests = [
      requestedRide.copyWith(
        id: 'other-ride',
        destination: const GeoPoint(latitude: 25, longitude: 68),
      ),
    ];
    await controller.load();
    await tester.pumpAndSettle();
    expect(controller.state.active?.id, 'other-ride');
    expect(tester.widget<RideMap>(find.byType(RideMap)).polylines, isEmpty);
    await tester.pumpWidget(const SizedBox.shrink());
  });
  for (final conflictCode in [
    'pricing_policy_changed',
    'suggested_fare_changed',
  ]) {
    testWidgets(
      'priced booking is read-only and explicitly resubmits after $conflictCode',
      (tester) async {
        final requests = _ConflictRequests(conflictCode);
        final routes = _PricedRoutes();
        await tester.pumpWidget(
          ProviderScope(
            overrides: [
              authRepositoryProvider.overrideWithValue(FakeAuthRepository()),
              capabilityStoreProvider.overrideWithValue(
                MemoryCapabilityStore(),
              ),
              rideRequestRepositoryProvider.overrideWithValue(requests),
              riderServiceRepositoryProvider.overrideWithValue(
                FakeRideServiceRepository(
                  services: const [
                    RideService(
                      code: 'economy',
                      displayName: 'Economy',
                      description: 'Standard',
                      displayOrder: 1,
                      presentationToken: 'car',
                      pricingRequired: true,
                    ),
                  ],
                ),
              ),
              deviceLocationProvider.overrideWithValue(
                const FakeDeviceLocation(),
              ),
              routePreviewRepositoryProvider.overrideWithValue(routes),
              rideFlowRepositoryProvider.overrideWithValue(
                FakeRideFlowRepository(),
              ),
            ],
            child: const MaterialApp(
              home: Scaffold(body: RiderRequestScreen()),
            ),
          ),
        );
        await tester.pumpAndSettle();
        await tester.drag(
          find.byKey(const Key('dashboardPanelDragHandle')),
          const Offset(0, -400),
        );
        await tester.pumpAndSettle();
        final container = ProviderScope.containerOf(
          tester.element(find.byType(RiderRequestScreen)),
        );
        final controller = container.read(riderRequestControllerProvider);
        controller.setPickup(requestedRide.pickup);
        controller.setDestination(requestedRide.destination);
        await tester.pumpAndSettle();
        final list = tester.widget<ListView>(find.byType(ListView).first);
        list.controller!.jumpTo(list.controller!.position.maxScrollExtent);
        await tester.pumpAndSettle();
        expect(
          tester
              .widget<TextField>(find.byKey(const Key('fareField')))
              .controller!
              .text,
          '125.00',
        );
        expect(routes.calls, 1);
        expect(
          tester.widget<TextField>(find.byKey(const Key('fareField'))).readOnly,
          isTrue,
        );
        await tester.pumpAndSettle();
        await tester.ensureVisible(find.byKey(const Key('requestRideButton')));
        await tester.pumpAndSettle();
        await tester.tap(
          find.byKey(const Key('requestRideButton')).hitTestable(),
        );
        await tester.pumpAndSettle();
        expect(routes.calls, 2);
        expect(requests.attempts, 1);
        expect(controller.state.active, isNull);
        expect(
          tester
              .widget<TextField>(find.byKey(const Key('fareField')))
              .controller!
              .text,
          '225.00',
        );
        expect(
          container
              .read(riderRoutePreviewControllerProvider)
              .state
              .preview!
              .pricingPolicyVersion,
          'v2',
        );
        expect(
          tester.widget<RideMap>(find.byType(RideMap)).polylines,
          hasLength(1),
        );
        await tester.ensureVisible(find.byKey(const Key('requestRideButton')));
        await tester.pumpAndSettle();
        await tester.tap(
          find.byKey(const Key('requestRideButton')).hitTestable(),
        );
        await tester.pumpAndSettle();
        expect(requests.attempts, 2);
        expect(requests.version, 'v2');
        expect(
          requests.submittedFare,
          const Money(amountMinor: 22500, currency: 'PKR'),
        );
        expect(controller.state.active?.id, 'ride-1');
        expect(
          tester.widget<RideMap>(find.byType(RideMap)).polylines,
          hasLength(1),
        );
        await tester.pumpWidget(const SizedBox.shrink());
      },
    );
  }

  testWidgets(
    'changing service clears the old proposal and fits the new suggestion',
    (tester) async {
      final routes = _PricedRoutes();
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWithValue(FakeAuthRepository()),
            capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
            rideRequestRepositoryProvider.overrideWithValue(
              FakeRideRequestRepository(),
            ),
            riderServiceRepositoryProvider.overrideWithValue(
              FakeRideServiceRepository(),
            ),
            deviceLocationProvider.overrideWithValue(
              const FakeDeviceLocation(),
            ),
            routePreviewRepositoryProvider.overrideWithValue(routes),
            rideFlowRepositoryProvider.overrideWithValue(
              FakeRideFlowRepository(),
            ),
          ],
          child: const MaterialApp(home: Scaffold(body: RiderRequestScreen())),
        ),
      );
      await tester.pumpAndSettle();
      final container = ProviderScope.containerOf(
        tester.element(find.byType(RiderRequestScreen)),
      );
      final controller = container.read(riderRequestControllerProvider);
      controller.setPickup(requestedRide.pickup);
      controller.setDestination(requestedRide.destination);
      await tester.pumpAndSettle();
      await tester.drag(
        find.byKey(const Key('dashboardPanelDragHandle')),
        const Offset(0, -400),
      );
      await tester.pumpAndSettle();
      final list = tester.widget<ListView>(find.byType(ListView).first);
      list.controller!.jumpTo(list.controller!.position.maxScrollExtent);
      await tester.pumpAndSettle();
      expect(
        tester.widget<TextField>(find.byKey(const Key('fareField'))).readOnly,
        isTrue,
      );
      await tester.pumpAndSettle();
      controller.selectService('comfort');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<TextField>(find.byKey(const Key('fareField')))
            .controller!
            .text,
        '225.00',
      );
      expect(
        tester.widget<TextField>(find.byKey(const Key('fareField'))).readOnly,
        isTrue,
      );
      await tester.pumpAndSettle();
      controller.setPickup(
        requestedRide.pickup,
        placeId: 'new-place-same-coordinates',
      );
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<TextField>(find.byKey(const Key('fareField')))
            .controller!
            .text,
        '225.00',
      );
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );
}

class _Routes implements RoutePreviewRepository {
  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  }) async => RoutePreview(
    distanceMeters: 2000,
    durationSeconds: 300,
    encodedPolyline: '_p~iF~ps|U_ulLnnqC',
  );

  @override
  void cancel() {}
}

class _PricedRoutes implements RoutePreviewRepository {
  int calls = 0;
  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  }) async {
    calls++;
    return RoutePreview(
      distanceMeters: 2000,
      durationSeconds: 300,
      encodedPolyline: '_p~iF~ps|U_ulLnnqC',
      suggestedFare: Money(
        amountMinor: calls == 1 ? 12500 : 22500,
        currency: 'PKR',
      ),
      pricingPolicyVersion: calls == 1 ? 'v1' : 'v2',
    );
  }

  @override
  void cancel() {}
}

class _ConflictRequests extends FakeRideRequestRepository {
  _ConflictRequests(this.conflictCode);
  final String conflictCode;
  int attempts = 0;
  String? version;
  @override
  Future<RideRequest> create({
    required GeoPoint pickup,
    required GeoPoint destination,
    required Money proposedFare,
    String serviceCode = 'economy',
    String? pricingPolicyVersion,
    String? pickupPlaceId,
    String? destinationPlaceId,
  }) async {
    attempts++;
    version = pricingPolicyVersion;
    if (attempts == 1) {
      throw ApiException(
        conflictCode,
        'Rates changed. Refresh the suggestion and submit again.',
        statusCode: 409,
      );
    }
    return super.create(
      pickup: pickup,
      destination: destination,
      proposedFare: proposedFare,
      serviceCode: serviceCode,
    );
  }
}
