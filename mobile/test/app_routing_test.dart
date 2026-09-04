import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';

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
}

Widget testApp(AuthRepository repository) => ProviderScope(
  overrides: [
    authRepositoryProvider.overrideWithValue(repository),
    capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
    rideRequestRepositoryProvider.overrideWithValue(
      FakeRideRequestRepository(requests: [requestedRide]),
    ),
    deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
  ],
  child: const UberCloneApp(),
);
