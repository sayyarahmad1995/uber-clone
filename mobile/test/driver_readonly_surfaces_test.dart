import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_onboarding.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('Driver drawer exposes real read-only destinations', (
    tester,
  ) async {
    await tester.pumpWidget(
      _testApp(
        accountHasDriver: true,
        driverRepository: FakeDriverRepository(profile: driverProfile),
        onboardingRepository: FakeDriverOnboardingRepository(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('capabilityMenuButton')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('drawerDriverDetails')), findsOneWidget);
    expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);

    await tester.tap(find.byKey(const Key('drawerDriverDetails')));
    await tester.pumpAndSettle();

    expect(find.text('Driver details'), findsOneWidget);
    expect(
      ModalRoute.of(tester.element(find.text('Driver details')))!.opaque,
      isFalse,
    );
    expect(find.text('Test Driver'), findsOneWidget);
    expect(find.text('Active'), findsOneWidget);
    expect(find.text('Offline'), findsOneWidget);
  });

  testWidgets(
    'approved onboarding snapshot backs Driver details and Vehicles',
    (tester) async {
      final application = DriverOnboardingApplication(
        id: 'application-approved',
        displayName: 'Approved Driver',
        status: 'approved',
        service: comfortService,
        vehicle: driverVehicle,
        submittedAt: DateTime.utc(2026, 9, 7),
        decidedAt: DateTime.utc(2026, 9, 8),
      );

      await tester.pumpWidget(
        _testApp(
          accountHasDriver: true,
          driverRepository: FakeDriverRepository(),
          onboardingRepository: FakeDriverOnboardingRepository(
            application: application,
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const Key('capabilityMenuButton')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('drawerDriverDetails')));
      await tester.pumpAndSettle();

      expect(find.text('Approved Driver'), findsOneWidget);
      expect(find.text('Approved'), findsOneWidget);
      expect(find.text('Comfort'), findsOneWidget);

      await tester.pageBack();
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('capabilityMenuButton')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('drawerVehicles')));
      await tester.pumpAndSettle();

      expect(find.text('Vehicles'), findsOneWidget);
      expect(ModalRoute.of(tester.element(find.text('Vehicles')))!.opaque, isFalse);
      expect(find.text('Toyota'), findsOneWidget);
      expect(find.text('Corolla'), findsOneWidget);
      expect(find.text('2024'), findsOneWidget);
      expect(find.text('ABC-123'), findsOneWidget);
      expect(find.text('Comfort'), findsOneWidget);
      expect(find.text('Driver onboarding application'), findsOneWidget);
    },
  );

  testWidgets('Rider-only drawer hides Driver read-only destinations', (
    tester,
  ) async {
    await tester.pumpWidget(
      _testApp(
        accountHasDriver: false,
        driverRepository: FakeDriverRepository(),
        onboardingRepository: FakeDriverOnboardingRepository(),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('capabilityMenuButton')));
    await tester.pumpAndSettle();

    expect(find.byKey(const Key('drawerDriverDetails')), findsNothing);
    expect(find.byKey(const Key('drawerVehicles')), findsNothing);
    expect(find.byKey(const Key('drawerBecomeDriver')), findsOneWidget);
  });
}

Widget _testApp({
  required bool accountHasDriver,
  required FakeDriverRepository driverRepository,
  required FakeDriverOnboardingRepository onboardingRepository,
}) {
  return ProviderScope(
    overrides: [
      authRepositoryProvider.overrideWithValue(
        FakeAuthRepository(
          account: accountHasDriver ? bothCapabilities : riderAccount,
        ),
      ),
      driverRepositoryProvider.overrideWithValue(driverRepository),
      driverOnboardingRepositoryProvider.overrideWithValue(
        onboardingRepository,
      ),
      capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
      rideRequestRepositoryProvider.overrideWithValue(
        FakeRideRequestRepository(requests: [requestedRide]),
      ),
      deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
    ],
    child: const UberCloneApp(),
  );
}
