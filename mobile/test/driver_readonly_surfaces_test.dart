import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_onboarding.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('Driver drawer stays open beneath Driver details', (
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

    final shellScaffold = tester.state<ScaffoldState>(
      find
          .ancestor(
            of: find.byKey(const Key('capabilityMenuButton')),
            matching: find.byType(Scaffold),
          )
          .first,
    );

    await tester.tap(find.byKey(const Key('capabilityMenuButton')));
    await tester.pumpAndSettle();

    expect(shellScaffold.isDrawerOpen, isTrue);
    expect(find.byKey(const Key('drawerDriverDetails')), findsOneWidget);
    expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);

    await tester.tap(find.byKey(const Key('drawerDriverDetails')));
    await tester.pumpAndSettle();

    final driverDetailsTitle = find.descendant(
      of: find.byType(AppBar),
      matching: find.text('Driver details'),
    );

    expect(shellScaffold.isDrawerOpen, isTrue);
    expect(find.text('Driver details'), findsNWidgets(2));
    expect(driverDetailsTitle, findsOneWidget);
    expect(ModalRoute.of(tester.element(driverDetailsTitle))!.opaque, isFalse);
    expect(find.text('Test Driver'), findsOneWidget);
    expect(find.text('Active'), findsOneWidget);
    expect(find.text('Offline'), findsOneWidget);

    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(shellScaffold.isDrawerOpen, isTrue);
    expect(find.byKey(const Key('drawerDriverDetails')), findsOneWidget);
    expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);
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

      final shellScaffold = tester.state<ScaffoldState>(
        find
            .ancestor(
              of: find.byKey(const Key('capabilityMenuButton')),
              matching: find.byType(Scaffold),
            )
            .first,
      );

      await tester.tap(find.byKey(const Key('capabilityMenuButton')));
      await tester.pumpAndSettle();
      expect(shellScaffold.isDrawerOpen, isTrue);

      await tester.tap(find.byKey(const Key('drawerDriverDetails')));
      await tester.pumpAndSettle();

      expect(shellScaffold.isDrawerOpen, isTrue);
      expect(find.text('Approved Driver'), findsOneWidget);
      expect(find.text('Approved'), findsOneWidget);
      expect(find.text('Comfort'), findsOneWidget);

      await tester.pageBack();
      await tester.pumpAndSettle();

      expect(shellScaffold.isDrawerOpen, isTrue);
      expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);

      await tester.tap(find.byKey(const Key('drawerVehicles')));
      await tester.pumpAndSettle();

      final vehiclesTitle = find.descendant(
        of: find.byType(AppBar),
        matching: find.text('Vehicles'),
      );

      expect(shellScaffold.isDrawerOpen, isTrue);
      expect(find.text('Vehicles'), findsNWidgets(2));
      expect(vehiclesTitle, findsOneWidget);
      expect(ModalRoute.of(tester.element(vehiclesTitle))!.opaque, isFalse);
      expect(find.text('Toyota'), findsOneWidget);
      expect(find.text('Corolla'), findsOneWidget);
      expect(find.text('2024'), findsOneWidget);
      expect(find.text('ABC-123'), findsOneWidget);
      expect(find.text('Comfort'), findsOneWidget);
      expect(find.text('Driver onboarding application'), findsOneWidget);

      await tester.pageBack();
      await tester.pumpAndSettle();

      expect(shellScaffold.isDrawerOpen, isTrue);
      expect(find.byKey(const Key('drawerDriverDetails')), findsOneWidget);
      expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);
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
