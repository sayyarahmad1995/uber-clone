import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
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
    controller.setPickup(const GeoPoint(latitude: 24.86, longitude: 67.01));
    controller.setDestination(
      const GeoPoint(latitude: 24.90, longitude: 67.05),
    );
    await tester.pumpAndSettle();

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
