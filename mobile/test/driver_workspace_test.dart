import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/driver_workspace/presentation/driver_workspace_screen.dart';

import 'app_routing_test.dart' show testApp;
import 'test_doubles.dart';

void main() {
  testWidgets('Rider can enable Driver access on the same account', (
    tester,
  ) async {
    await tester.pumpWidget(testApp(FakeAuthRepository(account: riderAccount)));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Become a Driver'));
    await tester.pumpAndSettle();
    expect(find.text('Driver dashboard'), findsOneWidget);
    await tester.drag(find.byKey(const Key('dashboardPanelDragHandle')), const Offset(0, -300));
    await tester.pumpAndSettle();
    expect(find.text('Driver setup'), findsOneWidget);
    expect(find.text('Become a Driver'), findsNothing);
  });

  testWidgets('Driver setup saves vehicle then resets to readiness panel', (
    tester,
  ) async {
    final repo = FakeDriverRepository();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          driverRepositoryProvider.overrideWithValue(repo),
          deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
        ],
        child: const MaterialApp(
          home: Scaffold(body: DriverWorkspaceScreen(accountID: 'user-1')),
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -300),
    );
    await tester.pumpAndSettle();
    final values = [
      'Test Driver',
      'Toyota',
      'Corolla',
      '2024',
      'White',
      'ABC-123',
    ];
    for (var i = 0; i < values.length; i++) {
      final field = find.byKey(ValueKey('driver-field-$i'));
      await tester.ensureVisible(field);
      await tester.pumpAndSettle();
      await tester.enterText(field, values[i]);
    }
    await tester.ensureVisible(find.text('Save Driver details'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Save Driver details'));
    await tester.pumpAndSettle();
    expect(repo.profile!.vehicle, driverVehicle);
    expect(find.text('Ready to go online'), findsOneWidget);
    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -300),
    );
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Go online'));
    await tester.tap(find.text('Go online'));
    await tester.pumpAndSettle();
    expect(repo.profile!.isOnline, isTrue);
    expect(find.text('Go offline'), findsOneWidget);
  });
}
