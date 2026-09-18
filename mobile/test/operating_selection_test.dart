import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/domain/driver_profile.dart';
import 'package:uber_clone/features/driver_workspace/domain/operating_state.dart';
import 'package:uber_clone/features/driver_workspace/domain/registered_vehicle.dart';
import 'package:uber_clone/features/driver_workspace/presentation/operating_selection.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('shows one selected vehicle with all eligible services', (
    tester,
  ) async {
    final repo = FakeDriverRepository(
      profile: driverProfile,
      vehicles: operatingVehicles,
    )..operation = const OperatingState(vehicleId: 'vehicle-a', valid: true);

    await _pump(tester, repo);

    expect(
      find.byKey(const ValueKey('operating-current-selection')),
      findsOneWidget,
    );
    expect(find.text('Current selection'), findsOneWidget);
    expect(find.text('Toyota Corolla 2024 • ABC-123'), findsWidgets);
    expect(find.text('Available services: Economy, Comfort'), findsWidgets);
    expect(
      find.byKey(const ValueKey('operating-option-vehicle-a')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey('operating-option-vehicle-b')),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey('operating-unavailable-vehicle-c')),
      findsOneWidget,
    );
    expect(find.text('No active approved services'), findsOneWidget);
  });

  testWidgets('selects another approved vehicle once', (tester) async {
    final repo = FakeDriverRepository(
      profile: driverProfile,
      vehicles: operatingVehicles,
    )..operation = const OperatingState(vehicleId: 'vehicle-a', valid: true);

    await _pump(tester, repo);
    await tester.tap(find.byKey(const ValueKey('operating-option-vehicle-b')));
    await tester.pumpAndSettle();

    expect(repo.operation.vehicleId, 'vehicle-b');
    expect(repo.operation.valid, isTrue);
  });

  testWidgets('locked Driver cannot change operating context', (tester) async {
    final repo =
        FakeDriverRepository(
            profile: driverProfile,
            vehicles: operatingVehicles,
          )
          ..operation = const OperatingState(
            vehicleId: 'vehicle-a',
            valid: true,
            canChange: false,
          );

    await _pump(tester, repo);
    await tester.tap(find.byKey(const ValueKey('operating-option-vehicle-b')));
    await tester.pumpAndSettle();

    expect(repo.operation.vehicleId, 'vehicle-a');
    expect(
      find.text(
        'Go offline and finish any active trip before changing selection.',
      ),
      findsOneWidget,
    );
  });
}

Future<void> _pump(WidgetTester tester, FakeDriverRepository repo) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        driverRepositoryProvider.overrideWithValue(repo),
        deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
      ],
      child: const MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(child: OperatingSelectionControl()),
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

final operatingVehicles = [
  RegisteredVehicle(
    id: 'vehicle-a',
    vehicle: const DriverVehicle(
      make: 'Toyota',
      model: 'Corolla',
      modelYear: 2024,
      color: 'White',
      licensePlate: 'ABC-123',
    ),
    enrollments: [
      ApprovedServiceEnrollment(
        serviceCode: 'economy',
        displayName: 'Economy',
        approvedAt: DateTime.utc(2026, 9, 10),
        serviceActive: true,
      ),
      ApprovedServiceEnrollment(
        serviceCode: 'comfort',
        displayName: 'Comfort',
        approvedAt: DateTime.utc(2026, 9, 11),
        serviceActive: true,
      ),
    ],
  ),
  RegisteredVehicle(
    id: 'vehicle-b',
    vehicle: const DriverVehicle(
      make: 'Honda',
      model: 'Civic',
      modelYear: 2025,
      color: 'Black',
      licensePlate: 'NEW-456',
    ),
    enrollments: [
      ApprovedServiceEnrollment(
        serviceCode: 'comfort',
        displayName: 'Comfort',
        approvedAt: DateTime.utc(2026, 9, 12),
        serviceActive: true,
      ),
    ],
  ),
  RegisteredVehicle(
    id: 'vehicle-c',
    vehicle: const DriverVehicle(
      make: 'Suzuki',
      model: 'Alto',
      modelYear: 2020,
      color: 'Silver',
      licensePlate: 'OLD-789',
    ),
    enrollments: [
      ApprovedServiceEnrollment(
        serviceCode: 'comfort',
        displayName: 'Comfort',
        approvedAt: DateTime.utc(2026, 9, 8),
        serviceActive: false,
      ),
    ],
  ),
];
