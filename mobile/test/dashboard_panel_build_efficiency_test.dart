import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';

void main() {
  testWidgets('direct drag does not rebuild panel content on every move', (
    tester,
  ) async {
    var panelBuilds = 0;

    await tester.pumpWidget(
      MaterialApp(
        home: RideDashboardScaffold(
          panelIdentity: 'performance',
          minPanelSize: 0.20,
          initialPanelSize: 0.20,
          maxPanelSize: 0.60,
          map: const ColoredBox(color: Colors.blueGrey),
          panelBuilder: (context, scrollController, scrollEnabled) {
            panelBuilds++;
            return ListView(
              controller: scrollController,
              physics: scrollEnabled
                  ? const ClampingScrollPhysics()
                  : const NeverScrollableScrollPhysics(),
              children: const [SizedBox(height: 600)],
            );
          },
        ),
      ),
    );
    await tester.pumpAndSettle();

    final buildsBeforeDrag = panelBuilds;
    final handle = find.byKey(const Key('dashboardPanelDragHandle'));
    final drag = await tester.startGesture(tester.getCenter(handle));

    for (var i = 0; i < 4; i++) {
      await drag.moveBy(const Offset(0, -30));
      await tester.pump();
    }

    expect(panelBuilds, buildsBeforeDrag);

    final panel = find.byKey(const Key('dashboardPanel'));
    final previewHeight = tester.getSize(panel).height;
    final bottomBeforeRelease = tester.getBottomLeft(panel).dy;
    final dashboardHeight = tester
        .getSize(find.byType(RideDashboardScaffold))
        .height;

    await drag.up();
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    final midTransitionHeight = tester.getSize(panel).height;
    expect(midTransitionHeight, greaterThan(previewHeight));
    expect(midTransitionHeight, lessThan(dashboardHeight * 0.60));
    expect(
      tester.getBottomLeft(panel).dy,
      moreOrLessEquals(bottomBeforeRelease),
    );

    await tester.pumpAndSettle();

    expect(
      tester.getSize(panel).height,
      moreOrLessEquals(dashboardHeight * 0.60),
    );
    expect(
      tester.getBottomLeft(panel).dy,
      moreOrLessEquals(bottomBeforeRelease),
    );
    expect(panelBuilds, buildsBeforeDrag + 1);
  });
}
