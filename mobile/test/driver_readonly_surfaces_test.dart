import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_onboarding.dart';
import 'package:uber_clone/features/driver_workspace/domain/operating_state.dart';
import 'package:uber_clone/features/driver_workspace/domain/registered_vehicle.dart';

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
    expect(find.byKey(const Key('drawerDriver')), findsOneWidget);
    expect(find.byKey(const Key('drawerDriverOnboarding')), findsNothing);
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
    'Driver capability without an application still shows Become a Driver',
    (tester) async {
      await tester.pumpWidget(
        _testApp(
          accountHasDriver: true,
          driverRepository: FakeDriverRepository(),
          onboardingRepository: FakeDriverOnboardingRepository(),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const Key('capabilityMenuButton')));
      await tester.pumpAndSettle();

      expect(find.byKey(const Key('drawerDriver')), findsNothing);
      expect(find.byKey(const Key('drawerDriverOnboarding')), findsNothing);
      expect(find.byKey(const Key('drawerBecomeDriver')), findsOneWidget);
      expect(find.text('Become a Driver'), findsOneWidget);
      expect(find.byKey(const Key('drawerDriverDetails')), findsNothing);
      expect(find.byKey(const Key('drawerVehicles')), findsNothing);

      await tester.tap(find.byKey(const Key('drawerBecomeDriver')));
      await tester.pumpAndSettle();

      expect(find.text('Become a Driver'), findsOneWidget);
    },
  );

  testWidgets(
    'pending Driver application keeps operational destinations hidden',
    (tester) async {
      final driverRepository = FakeDriverRepository();
      final application = DriverOnboardingApplication(
        id: 'application-pending',
        displayName: 'Pending Driver',
        status: 'pending',
        service: comfortService,
        vehicle: driverVehicle,
        submittedAt: DateTime.utc(2026, 9, 17),
      );

      await tester.pumpWidget(
        _testApp(
          accountHasDriver: true,
          driverRepository: driverRepository,
          onboardingRepository: FakeDriverOnboardingRepository(
            application: application,
          ),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const Key('capabilityMenuButton')));
      await tester.pumpAndSettle();

      expect(find.byKey(const Key('drawerRider')), findsOneWidget);
      expect(find.byKey(const Key('drawerDriver')), findsOneWidget);
      expect(find.byKey(const Key('drawerDriverOnboarding')), findsNothing);
      expect(find.byKey(const Key('drawerBecomeDriver')), findsNothing);
      expect(find.byKey(const Key('drawerDriverDetails')), findsNothing);
      expect(find.byKey(const Key('drawerVehicles')), findsNothing);

      await tester.tap(find.byKey(const Key('drawerDriver')));
      await tester.pumpAndSettle();

      expect(find.text('Application under review'), findsOneWidget);

      driverRepository.profile = driverProfile;
      await tester.scrollUntilVisible(
        find.text('Refresh application'),
        200,
        scrollable: find.byType(Scrollable).last,
      );
      await tester.tap(find.text('Refresh application'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const Key('capabilityMenuButton')));
      await tester.pumpAndSettle();

      expect(find.byKey(const Key('drawerDriver')), findsOneWidget);
      expect(find.byKey(const Key('drawerDriverOnboarding')), findsNothing);
      expect(find.byKey(const Key('drawerDriverDetails')), findsOneWidget);
      expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);
    },
  );

  testWidgets('rejected Driver application remains accessible as Driver', (
    tester,
  ) async {
    final application = DriverOnboardingApplication(
      id: 'application-rejected',
      displayName: 'Rejected Driver',
      status: 'rejected',
      service: comfortService,
      vehicle: driverVehicle,
      rejectionReason: 'Document review failed.',
      submittedAt: DateTime.utc(2026, 9, 17),
      decidedAt: DateTime.utc(2026, 9, 18),
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

    expect(find.byKey(const Key('drawerDriver')), findsOneWidget);
    expect(find.byKey(const Key('drawerBecomeDriver')), findsNothing);
    expect(find.byKey(const Key('drawerDriverDetails')), findsNothing);
    expect(find.byKey(const Key('drawerVehicles')), findsNothing);

    await tester.tap(find.byKey(const Key('drawerDriver')));
    await tester.pumpAndSettle();

    expect(find.text('Application rejected'), findsOneWidget);
  });

  testWidgets('approved profile exposes Driver details and Vehicles', (
    tester,
  ) async {
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
        driverRepository: FakeDriverRepository(
          profile: driverProfile.copyWith(
            displayName: 'Approved Driver',
            status: 'approved',
          ),
          vehicles: selectedCorollaVehicles,
        )..operation = const OperatingState(vehicleId: 'corolla', valid: true),
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
    expect(find.text('Selected vehicle'), findsOneWidget);
    expect(find.text('Toyota Corolla'), findsOneWidget);
    expect(find.text('Available services'), findsOneWidget);
    expect(find.text('Economy, Comfort'), findsOneWidget);
    expect(find.text('Selected service'), findsNothing);

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
    expect(find.text('Toyota Corolla'), findsOneWidget);
    expect(find.text('2024'), findsOneWidget);
    expect(find.text('ABC-123'), findsOneWidget);
    expect(find.text('Economy'), findsOneWidget);
    expect(find.text('Comfort'), findsOneWidget);
    expect(find.text('Driver onboarding application'), findsNothing);

    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(shellScaffold.isDrawerOpen, isTrue);
    expect(find.byKey(const Key('drawerDriverDetails')), findsOneWidget);
    expect(find.byKey(const Key('drawerVehicles')), findsOneWidget);
  });

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

final selectedCorollaVehicles = [
  RegisteredVehicle(
    id: 'corolla',
    vehicle: driverVehicle,
    enrollments: [
      ApprovedServiceEnrollment(
        serviceCode: 'economy',
        displayName: 'Economy',
        approvedAt: DateTime.utc(2026, 9, 8),
        serviceActive: true,
      ),
      ApprovedServiceEnrollment(
        serviceCode: 'comfort',
        displayName: 'Comfort',
        approvedAt: DateTime.utc(2026, 9, 9),
        serviceActive: true,
      ),
    ],
  ),
];

Widget _testApp({
  required bool accountHasDriver,
  required FakeDriverRepository driverRepository,
  required FakeDriverOnboardingRepository onboardingRepository,
}) {
  return ProviderScope(
    overrides: [
      rideFlowRepositoryProvider.overrideWithValue(FakeRideFlowRepository()),
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
