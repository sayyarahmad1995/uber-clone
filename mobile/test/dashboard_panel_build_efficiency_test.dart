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

    await drag.up();
    await tester.pumpAndSettle();

    expect(panelBuilds, buildsBeforeDrag + 1);
  });

  testWidgets('collapse snap uses paint-only intermediate motion', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: RideDashboardScaffold(
          panelIdentity: 'snap-animation',
          minPanelSize: 0.20,
          initialPanelSize: 0.20,
          maxPanelSize: 0.60,
          map: const ColoredBox(color: Colors.blueGrey),
          panelBuilder: (context, scrollController, scrollEnabled) => ListView(
            controller: scrollController,
            physics: scrollEnabled
                ? const ClampingScrollPhysics()
                : const NeverScrollableScrollPhysics(),
            children: const [SizedBox(height: 600)],
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final dashboard = find.byType(RideDashboardScaffold);
    final panel = find.byKey(const Key('dashboardPanel'));
    final dashboardHeight = tester.getSize(dashboard).height;
    final dashboardBottom = tester.getBottomLeft(dashboard).dy;
    final minHeight = dashboardHeight * 0.20;
    final minTop = dashboardBottom - minHeight;

    await tester.drag(
      find.byKey(const Key('dashboardPanelDragHandle')),
      const Offset(0, -300),
    );
    await tester.pumpAndSettle();

    final handle = find.byKey(const Key('dashboardPanelDragHandle'));
    final drag = await tester.startGesture(tester.getCenter(handle));
    await drag.moveBy(const Offset(0, 120));
    await tester.pump();
    final previewHeight = tester.getSize(panel).height;
    final previewTop = tester.getTopLeft(panel).dy;
    final remainingDistance = minTop - previewTop;

    await drag.up();
    await tester.pump();
    expect(tester.getTopLeft(panel).dy, closeTo(previewTop, 0.5));

    await tester.pump(const Duration(milliseconds: 16));
    final firstFrameTop = tester.getTopLeft(panel).dy;
    final firstFrameTravel = firstFrameTop - previewTop;
    expect(firstFrameTravel, greaterThanOrEqualTo(0));
    expect(firstFrameTravel, lessThan(remainingDistance * 0.08));

    await tester.pump(const Duration(milliseconds: 64));
    final midAnimationHeight = tester.getSize(panel).height;
    final midAnimationTop = tester.getTopLeft(panel).dy;

    expect(midAnimationHeight, closeTo(previewHeight, 0.5));
    expect(midAnimationTop, greaterThan(previewTop + 1));
    expect(midAnimationTop, lessThan(minTop - 1));

    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, closeTo(minHeight, 0.5));
    expect(tester.getTopLeft(panel).dy, closeTo(minTop, 0.5));
  });
}
