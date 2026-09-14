import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/data/driver_onboarding_repository.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_onboarding.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_profile.dart';
import 'package:uber_clone/features/driver_workspace/presentation/driver_workspace_screen.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('approved Driver submits an additional vehicle service application', (
    tester,
  ) async {
    final onboarding = AdditionalApplicationRepository(
      application: initialApprovedApplication,
    );
    await _pumpDriverWorkspace(tester, onboarding);
    await _expandPanel(tester);

    expect(find.text('Vehicle/service applications'), findsOneWidget);
    expect(find.text('Add vehicle/service'), findsOneWidget);

    await tester.ensureVisible(
      find.byKey(const Key('driverAddVehicleServiceApplicationButton')),
    );
    await tester.tap(
      find.byKey(const Key('driverAddVehicleServiceApplicationButton')),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('additional-service-field')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Comfort').last);
    await tester.pumpAndSettle();

    final values = ['Honda', 'Civic', '2025', 'Black', 'NEW-456'];
    for (var i = 0; i < values.length; i++) {
      final field = find.byKey(ValueKey('additional-vehicle-field-$i'));
      await tester.ensureVisible(field);
      await tester.pumpAndSettle();
      await tester.enterText(field, values[i]);
    }

    await tester.ensureVisible(find.text('Review application'));
    await tester.tap(find.text('Review application'));
    await tester.pumpAndSettle();

    expect(onboarding.calls, contains('precheck:comfort'));
    expect(find.text('Review vehicle/service application'), findsOneWidget);
    expect(find.text('Service: Comfort'), findsOneWidget);

    await tester.tap(find.text('Submit for review'));
    await tester.pumpAndSettle();

    expect(onboarding.calls, contains('submit:comfort'));
    expect(
      onboarding.application!.applicationType,
      'additional_vehicle_service',
    );
    expect(onboarding.application!.isPending, isTrue);
    expect(onboarding.application!.vehicle.licensePlate, 'NEW-456');
  });

  testWidgets('pending additional application disables another submission', (
    tester,
  ) async {
    await _pumpDriverWorkspace(
      tester,
      AdditionalApplicationRepository(application: pendingAdditionalApplication),
    );
    await _expandPanel(tester);

    await tester.ensureVisible(
      find.byKey(const Key('driverAddVehicleServiceApplicationButton')),
    );
    final addButton = tester.widget<FilledButton>(
      find.byKey(const Key('driverAddVehicleServiceApplicationButton')),
    );

    expect(
      find.text('Latest additional application: Under review'),
      findsOneWidget,
    );
    expect(addButton.onPressed, isNull);
  });
}

Future<void> _pumpDriverWorkspace(
  WidgetTester tester,
  DriverOnboardingRepository onboarding,
) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        rideFlowRepositoryProvider.overrideWithValue(FakeRideFlowRepository()),
        driverRepositoryProvider.overrideWithValue(
          FakeDriverRepository(profile: driverProfile),
        ),
        driverOnboardingRepositoryProvider.overrideWithValue(onboarding),
        deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
      ],
      child: const MaterialApp(
        home: Scaffold(body: DriverWorkspaceScreen(accountID: 'user-1')),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

Future<void> _expandPanel(WidgetTester tester) async {
  final panel = find.byKey(const Key('dashboardPanel'));
  final dashboardHeight = tester
      .getSize(find.byType(RideDashboardScaffold))
      .height;
  await tester.drag(
    find.byKey(const Key('dashboardPanelDragHandle')),
    const Offset(0, -300),
  );
  await tester.pumpAndSettle();
  expect(tester.getSize(panel).height, dashboardHeight * 0.60);
}

final initialApprovedApplication = DriverOnboardingApplication(
  id: 'initial-application',
  applicationType: 'initial_onboarding',
  displayName: 'Test Driver',
  status: 'approved',
  service: economyService,
  vehicle: driverVehicle,
  submittedAt: DateTime.utc(2026, 9, 13),
  decidedAt: DateTime.utc(2026, 9, 13),
);

final pendingAdditionalApplication = DriverOnboardingApplication(
  id: 'additional-application',
  applicationType: 'additional_vehicle_service',
  displayName: 'Test Driver',
  status: 'pending',
  service: comfortService,
  vehicle: const DriverVehicle(
    make: 'Honda',
    model: 'Civic',
    modelYear: 2025,
    color: 'Black',
    licensePlate: 'NEW-456',
  ),
  submittedAt: DateTime.utc(2026, 9, 14),
);

class AdditionalApplicationRepository implements DriverOnboardingRepository {
  AdditionalApplicationRepository({this.application});

  DriverOnboardingApplication? application;
  final calls = <String>[];

  @override
  Future<List<DriverServiceOption>> listServices() async => const [
    economyService,
    comfortService,
  ];

  @override
  Future<DriverOnboardingApplication?> getLatest() async => application;

  @override
  Future<DriverOnboardingPrecheck> precheck({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async {
    calls.add('precheck:$serviceCode');
    return DriverOnboardingPrecheck(
      eligible: true,
      reasons: const [],
      service: _service(serviceCode),
    );
  }

  @override
  Future<DriverOnboardingApplication> submit({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async {
    calls.add('submit:$serviceCode');
    return application = DriverOnboardingApplication(
      id: 'submitted-additional-application',
      applicationType: 'additional_vehicle_service',
      displayName: displayName,
      status: 'pending',
      service: _service(serviceCode),
      vehicle: vehicle,
      submittedAt: DateTime.utc(2026, 9, 14),
    );
  }

  DriverServiceOption _service(String code) => switch (code) {
    'comfort' => comfortService,
    _ => economyService,
  };
}
