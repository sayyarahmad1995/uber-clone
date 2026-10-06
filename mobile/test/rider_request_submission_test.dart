import 'package:flutter/material.dart';
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

    expect(tester.widget<RideMap>(find.byType(RideMap)).polylines, hasLength(1));

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
    expect(bookedMap.polylines.single.points, hasLength(3));
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
