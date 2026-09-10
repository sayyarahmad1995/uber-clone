import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/domain/registered_vehicle.dart';
import 'package:uber_clone/features/driver_workspace/presentation/driver_readonly_surfaces.dart';

Map<String, dynamic> vehicleJson(String id, List<Map<String, dynamic>> services) => {
  'id': id, 'make': 'Toyota', 'model': 'Corolla', 'model_year': null,
  'color': 'White', 'license_plate': id, 'approved_service_enrollments': services,
};

void main() {
  final comfort = {
    'service_code': 'comfort', 'display_name': 'Comfort',
    'approved_at': '2026-09-10T00:00:00Z', 'service_active': false,
  };
  test('vehicle preserves identity and explicit enrollments with nullable year', () {
    final record = RegisteredVehicle.fromJson(vehicleJson('one', [comfort]));
    expect(record.id, 'one');
    expect(record.vehicle.modelYear, isNull);
    expect(record.enrollments.single.serviceCode, 'comfort');
    expect(record.enrollments.single.serviceActive, isFalse);
    expect(RegisteredVehicle.fromJson(vehicleJson('two', [])).enrollments, isEmpty);
  });
  testWidgets('Vehicles renders multiple records and exact service availability', (tester) async {
    await tester.pumpWidget(ProviderScope(
      overrides: [driverVehiclesProvider.overrideWith((ref) async => [
        RegisteredVehicle.fromJson(vehicleJson('one', [comfort])),
        RegisteredVehicle.fromJson(vehicleJson('two', [])),
      ])],
      child: const MaterialApp(home: DriverVehiclesScreen()),
    ));
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('one')), findsOneWidget);
    expect(find.byKey(const ValueKey('two')), findsOneWidget);
    expect(find.text('Comfort (currently unavailable)'), findsOneWidget);
    expect(find.text('Economy'), findsNothing);
    await tester.scrollUntilVisible(find.text('No approved service enrollments'), 200);
    expect(find.text('No approved service enrollments'), findsOneWidget);
  });
  testWidgets('Vehicles shows retry when authoritative read fails', (tester) async {
    await tester.pumpWidget(ProviderScope(
      overrides: [driverVehiclesProvider.overrideWith((ref) async => throw Exception('offline'))],
      child: const MaterialApp(home: DriverVehiclesScreen()),
    ));
    await tester.pumpAndSettle();
    expect(find.text('Unable to load vehicles'), findsOneWidget);
    expect(find.text('Retry'), findsOneWidget);
  });
}
