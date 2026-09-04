import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
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
    expect(find.text('Driver workspace'), findsOneWidget);
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
    final sheet = tester.widget<DraggableScrollableSheet>(
      find.byType(DraggableScrollableSheet),
    );
    expect(sheet.initialChildSize, 0.50);
    expect(sheet.maxChildSize, 0.70);

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
