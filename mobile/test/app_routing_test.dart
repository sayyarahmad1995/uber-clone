import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/data/ride_request_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('signed-out startup routes to application login', (tester) async {
    await tester.pumpWidget(testApp(FakeAuthRepository()));
    await tester.pumpAndSettle();
    expect(find.text('Welcome back'), findsOneWidget);
    expect(find.text('Sign in'), findsOneWidget);
  });

  testWidgets('restored account enters Rider by default', (tester) async {
    await tester.pumpWidget(testApp(FakeAuthRepository(account: riderAccount)));
    await tester.pumpAndSettle();
    expect(find.text('Looking for Driver offers'), findsOneWidget);
    expect(find.text('Rider'), findsOneWidget);
  });

  testWidgets('dual-capability account can switch to Driver', (tester) async {
    await tester.pumpWidget(
      testApp(FakeAuthRepository(account: bothCapabilities)),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Driver'));
    await tester.pumpAndSettle();
    expect(find.text('Driver dashboard'), findsOneWidget);
  });

  testWidgets('Rider dashboard focuses on device location at startup', (
    tester,
  ) async {
    final location = CountingDeviceLocation();
    await tester.pumpWidget(
      testApp(
        FakeAuthRepository(account: riderAccount),
        rideRequests: FakeRideRequestRepository(),
        deviceLocation: location,
      ),
    );
    await tester.pumpAndSettle();

    expect(location.requestCount, 1);
    final panel = find.byKey(const Key('dashboardPanel'));
    final dashboardHeight = tester
        .getSize(find.byType(RideDashboardScaffold))
        .height;
    final collapsedHeight = tester.getSize(panel).height;
    expect(collapsedHeight, moreOrLessEquals(dashboardHeight * 0.18));

    await tester.drag(find.text('Where are you going?'), const Offset(0, 80));
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, collapsedHeight);

    final maximizeGesture = await tester.startGesture(
      tester.getCenter(find.text('Where are you going?')),
    );
    await maximizeGesture.moveBy(const Offset(0, -100));
    await tester.pump();
    expect(tester.getSize(panel).height, collapsedHeight);

    await maximizeGesture.up();
    await tester.pumpAndSettle();
    final expandedHeight = tester.getSize(panel).height;
    expect(expandedHeight, moreOrLessEquals(dashboardHeight * 0.60));

    final panelList = find.descendant(
      of: panel,
      matching: find.byType(ListView),
    );
    await tester.drag(panelList, const Offset(0, -300));
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, expandedHeight);

    await tester.drag(panelList, const Offset(0, 600));
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, expandedHeight);

    final minimizeGesture = await tester.startGesture(
      tester.getCenter(panelList),
    );
    await minimizeGesture.moveBy(const Offset(0, 100));
    await tester.pump();
    expect(tester.getSize(panel).height, expandedHeight);

    await minimizeGesture.up();
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, collapsedHeight);

    await tester.tap(find.byTooltip('Center map on your location'));
    await tester.pumpAndSettle();

    expect(location.requestCount, 2);
  });
}

Widget testApp(
  AuthRepository repository, {
  RideRequestRepository? rideRequests,
  DeviceLocation? deviceLocation,
}) => ProviderScope(
  overrides: [
    authRepositoryProvider.overrideWithValue(repository),
    capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
    rideRequestRepositoryProvider.overrideWithValue(
      rideRequests ?? FakeRideRequestRepository(requests: [requestedRide]),
    ),
    deviceLocationProvider.overrideWithValue(
      deviceLocation ?? const FakeDeviceLocation(),
    ),
  ],
  child: const UberCloneApp(),
);

class CountingDeviceLocation implements DeviceLocation {
  int requestCount = 0;

  @override
  Future<GeoPoint> current() async {
    requestCount++;
    return const GeoPoint(latitude: 24.86, longitude: 67.01);
  }
}
