import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';
import 'package:uber_clone/core/providers.dart';

import 'app_routing_test.dart' show riderAccount, testApp;
import 'test_doubles.dart';

void main() {
  testWidgets('panel session changes do not rebuild the capability shell', (
    tester,
  ) async {
    await tester.pumpWidget(testApp(FakeAuthRepository(account: riderAccount)));
    await tester.pumpAndSettle();

    final shellScaffold = find
        .ancestor(
          of: find.byKey(const Key('capabilityMenuButton')),
          matching: find.byType(Scaffold),
        )
        .first;
    final before = tester.widget<Scaffold>(shellScaffold);
    final panel = find.byKey(const Key('dashboardPanel'));
    final dashboardHeight = tester
        .getSize(find.byType(RideDashboardScaffold))
        .height;

    final container = ProviderScope.containerOf(
      tester.element(find.byType(RideDashboardScaffold)),
    );
    container.read(dashboardPanelSessionProvider).setExpanded(true);
    await tester.pump();

    final afterExpand = tester.widget<Scaffold>(shellScaffold);
    expect(identical(afterExpand, before), isTrue);
    expect(tester.getSize(panel).height, closeTo(dashboardHeight * 0.60, 0.5));

    container.read(dashboardPanelSessionProvider).setExpanded(false);
    await tester.pump();

    final afterCollapse = tester.widget<Scaffold>(shellScaffold);
    expect(identical(afterCollapse, before), isTrue);
    expect(tester.getSize(panel).height, closeTo(dashboardHeight * 0.18, 0.5));
  });
}
