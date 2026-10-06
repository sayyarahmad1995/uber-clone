import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/maps/ride_map.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/rider_request/data/ride_service_repository.dart';
import 'package:uber_clone/features/rider_request/data/route_preview_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';
import 'package:uber_clone/features/rider_request/domain/ride_service.dart';
import 'package:uber_clone/features/rider_request/domain/route_preview.dart';
import 'package:uber_clone/features/rider_request/presentation/rider_request_screen.dart';

import 'test_doubles.dart';

const _third = RideService(
  code: 'third_car',
  displayName: 'Third category',
  description: 'A newly published category',
  displayOrder: 30,
  presentationToken: 'future-icon',
);

void main() {
  testWidgets('booking starts with service cards and requires a selection', (
    tester,
  ) async {
    await _pumpDashboard(tester);
    expect(find.text('Choose your ride'), findsOneWidget);
    expect(find.byKey(const Key('ride-service-card-economy')), findsOneWidget);
    expect(find.byType(DropdownButtonFormField<String>), findsNothing);
    expect(find.byKey(const Key('pickupSearchField')), findsNothing);
    expect(find.byKey(const Key('fareField')), findsNothing);
    expect(find.byKey(const Key('requestRideButton')), findsNothing);
    final chooserMap = tester.widget<RideMap>(find.byType(RideMap));
    expect(chooserMap.onTap, isNull);
    expect(chooserMap.onPlaceTap, isNull);
    expect(chooserMap.showCenterPin, isFalse);

    await tester.tap(find.byKey(const Key('ride-service-card-economy')));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('selectedRideService')), findsOneWidget);
    expect(find.byKey(const Key('changeRideServiceButton')), findsOneWidget);
    final bookingMap = tester.widget<RideMap>(find.byType(RideMap));
    expect(bookingMap.onTap, isNotNull);
    expect(bookingMap.onPlaceTap, isNotNull);
    await tester.scrollUntilVisible(
      find.byKey(const Key('pickupSearchField')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.byKey(const Key('pickupSearchField')), findsOneWidget);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('catalog refresh adds a selectable future service card', (
    tester,
  ) async {
    final catalog = _Catalog(FakeRideServiceRepository().services);
    final container = await _pumpDashboard(tester, catalog: catalog);
    catalog.services = [...catalog.services, _third];
    container.invalidate(riderServicesProvider);
    await tester.pumpAndSettle();
    final list = tester.widget<ListView>(find.byType(ListView).first);
    list.controller!.jumpTo(list.controller!.position.maxScrollExtent);
    await tester.pumpAndSettle();
    await tester.ensureVisible(
      find.byKey(const Key('ride-service-card-third_car')),
    );
    await tester.tap(
      find.byKey(const Key('ride-service-card-third_car')).hitTestable(),
    );
    await tester.pumpAndSettle();

    expect(
      container.read(riderRequestControllerProvider).serviceCode,
      'third_car',
    );
    expect(find.text('Third category'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.byKey(const Key('pickupSearchField')),
      100,
      scrollable: find.byType(Scrollable).first,
    );
    expect(find.byKey(const Key('pickupSearchField')), findsOneWidget);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('changing service preserves locations and refreshes the route', (
    tester,
  ) async {
    final routes = _Routes();
    final container = await _pumpDashboard(tester, routes: routes);
    await tester.tap(find.byKey(const Key('ride-service-card-economy')));
    await tester.pumpAndSettle();
    final controller = container.read(riderRequestControllerProvider);
    const pickup = GeoPoint(latitude: 24.86, longitude: 67.01);
    const destination = GeoPoint(latitude: 24.90, longitude: 67.05);
    controller.setPickup(pickup, placeId: 'pickup-place');
    controller.setDestination(destination, placeId: 'destination-place');
    await tester.pumpAndSettle();
    final list = tester.widget<ListView>(find.byType(ListView).first);
    list.controller!.jumpTo(list.controller!.position.maxScrollExtent);
    await tester.pumpAndSettle();
    await tester.enterText(find.byKey(const Key('fareField')), '735');
    FocusManager.instance.primaryFocus?.unfocus();
    await tester.pumpAndSettle();
    list.controller!.jumpTo(list.controller!.position.minScrollExtent);
    await tester.pumpAndSettle();
    await tester.ensureVisible(
      find.byKey(const Key('changeRideServiceButton')),
    );
    await tester.pumpAndSettle();
    await tester.tap(
      find.byKey(const Key('changeRideServiceButton')).hitTestable(),
    );
    await tester.pumpAndSettle();
    expect(find.text('Choose your ride'), findsOneWidget);
    expect(find.byKey(const Key('pickupSearchField')), findsNothing);
    expect(controller.state.pickup, pickup);
    expect(controller.state.destination, destination);

    await tester.ensureVisible(
      find.byKey(const Key('ride-service-card-comfort')),
    );
    await tester.tap(
      find.byKey(const Key('ride-service-card-comfort')).hitTestable(),
    );
    await tester.pumpAndSettle();
    expect(controller.serviceCode, 'comfort');
    expect(controller.state.pickupPlaceId, 'pickup-place');
    expect(controller.state.destinationPlaceId, 'destination-place');
    expect(routes.services, ['economy', 'comfort']);
    final bookingList = tester.widget<ListView>(find.byType(ListView).first);
    bookingList.controller!.jumpTo(
      bookingList.controller!.position.maxScrollExtent,
    );
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<TextField>(find.byKey(const Key('fareField')))
          .controller!
          .text,
      '735',
    );
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('deactivating the selected service returns to service cards', (
    tester,
  ) async {
    final catalog = _Catalog(FakeRideServiceRepository().services);
    final container = await _pumpDashboard(tester, catalog: catalog);
    await tester.tap(find.byKey(const Key('ride-service-card-economy')));
    await tester.pumpAndSettle();
    final controller = container.read(riderRequestControllerProvider);
    const pickup = GeoPoint(latitude: 24.86, longitude: 67.01);
    controller.setPickup(pickup);
    catalog.services = catalog.services
        .where((s) => s.code == 'comfort')
        .toList();
    container.invalidate(riderServicesProvider);
    await tester.pumpAndSettle();

    expect(find.text('Choose your ride'), findsOneWidget);
    expect(find.byKey(const Key('ride-service-card-economy')), findsNothing);
    expect(find.byKey(const Key('selectedRideService')), findsNothing);
    expect(controller.serviceCode, isEmpty);
    expect(controller.state.pickup, pickup);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('empty catalog offers retry without a booking form', (
    tester,
  ) async {
    final catalog = _Catalog([]);
    final container = await _pumpDashboard(tester, catalog: catalog);
    expect(find.text('No ride services currently available.'), findsOneWidget);
    expect(find.byKey(const Key('pickupSearchField')), findsNothing);
    catalog.services = FakeRideServiceRepository().services;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('ride-service-card-economy')), findsOneWidget);
    expect(container.read(riderRequestControllerProvider).state.active, isNull);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets('catalog failure can be retried before selecting a service', (
    tester,
  ) async {
    final catalog = _Catalog(FakeRideServiceRepository().services)..fail = true;
    await _pumpDashboard(tester, catalog: catalog);
    expect(find.text('Unable to load ride services.'), findsOneWidget);
    expect(find.byKey(const Key('pickupSearchField')), findsNothing);
    catalog.fail = false;
    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();
    expect(find.byKey(const Key('ride-service-card-economy')), findsOneWidget);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  testWidgets(
    'controller replacement resets service choice and booking inputs',
    (tester) async {
      final container = await _pumpDashboard(tester);
      await tester.tap(find.byKey(const Key('ride-service-card-economy')));
      await tester.pumpAndSettle();
      container
          .read(riderRequestControllerProvider)
          .setPickup(const GeoPoint(latitude: 24.86, longitude: 67.01));
      container.invalidate(riderRequestControllerProvider);
      await tester.pumpAndSettle();
      expect(find.text('Choose your ride'), findsOneWidget);
      expect(find.byKey(const Key('selectedRideService')), findsNothing);
      expect(
        container.read(riderRequestControllerProvider).state.pickup,
        isNull,
      );
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );
}

Future<ProviderContainer> _pumpDashboard(
  WidgetTester tester, {
  _Catalog? catalog,
  _Routes? routes,
}) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        authRepositoryProvider.overrideWithValue(FakeAuthRepository()),
        capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
        rideRequestRepositoryProvider.overrideWithValue(
          FakeRideRequestRepository(),
        ),
        riderServiceRepositoryProvider.overrideWithValue(
          catalog ?? _Catalog(FakeRideServiceRepository().services),
        ),
        deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
        routePreviewRepositoryProvider.overrideWithValue(routes ?? _Routes()),
        rideFlowRepositoryProvider.overrideWithValue(FakeRideFlowRepository()),
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
  return ProviderScope.containerOf(
    tester.element(find.byType(RiderRequestScreen)),
  );
}

class _Catalog implements RideServiceRepository {
  _Catalog(this.services);
  List<RideService> services;
  bool fail = false;

  @override
  Future<List<RideService>> list() async {
    if (fail) throw Exception('Catalog unavailable');
    return List.of(services);
  }
}

class _Routes implements RoutePreviewRepository {
  final services = <String>[];

  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  }) async {
    services.add(serviceCode);
    return RoutePreview(
      distanceMeters: 2000,
      durationSeconds: 300,
      encodedPolyline: '_p~iF~ps|U_ulLnnqC',
    );
  }

  @override
  void cancel() {}
}
