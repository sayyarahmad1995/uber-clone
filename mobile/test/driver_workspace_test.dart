import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
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
    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -300),
    );
    await tester.pumpAndSettle();
    expect(find.text('Become a Driver'), findsOneWidget);
    expect(
      find.text('Choose one service for your first vehicle application.'),
      findsOneWidget,
    );
  });

  testWidgets('Driver onboarding submits selected service for review', (
    tester,
  ) async {
    final driverRepo = FakeDriverRepository();
    final onboardingRepo = FakeDriverOnboardingRepository();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          driverRepositoryProvider.overrideWithValue(driverRepo),
          driverOnboardingRepositoryProvider.overrideWithValue(onboardingRepo),
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

    await tester.tap(find.byKey(const Key('driver-service-field')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Comfort').last);
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

    await tester.ensureVisible(find.text('Review application'));
    await tester.tap(find.text('Review application'));
    await tester.pumpAndSettle();
    expect(onboardingRepo.calls, contains('precheck:comfort'));
    expect(find.byType(AlertDialog), findsOneWidget);
    expect(find.text('Service: Comfort'), findsOneWidget);

    await tester.tap(find.text('Submit for review'));
    await tester.pumpAndSettle();

    expect(onboardingRepo.calls, contains('submit:comfort'));
    expect(driverRepo.profile, isNull);
    expect(find.text('Application under review'), findsOneWidget);

    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -300),
    );
    await tester.pumpAndSettle();
    expect(find.text('Service: Comfort'), findsOneWidget);
  });

  testWidgets('existing operational Driver can still go online', (
    tester,
  ) async {
    final repo = FakeDriverRepository(profile: driverProfile);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          driverRepositoryProvider.overrideWithValue(repo),
          driverOnboardingRepositoryProvider.overrideWithValue(
            FakeDriverOnboardingRepository(),
          ),
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
    await tester.ensureVisible(find.text('Go online'));
    await tester.tap(find.text('Go online'));
    await tester.pumpAndSettle();
    expect(repo.calls.take(2).toList(), ['location', 'online=true']);
    expect(repo.profile!.isOnline, isTrue);
    expect(find.text('Go offline'), findsOneWidget);
  });
}
