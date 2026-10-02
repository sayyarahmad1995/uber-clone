import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';

void main() {
  testWidgets(
    'touch jitter at top does not steal an upward content scroll',
    (tester) async {
      ScrollController? contentController;

      await tester.pumpWidget(
        MaterialApp(
          home: RideDashboardScaffold(
            panelIdentity: 'jitter-handoff',
            minPanelSize: 0.20,
            initialPanelSize: 0.60,
            maxPanelSize: 0.60,
            map: const ColoredBox(color: Colors.blueGrey),
            panelBuilder: (context, scrollController, scrollEnabled) {
              contentController = scrollController;
              return ListView(
                key: const Key('jitterList'),
                controller: scrollController,
                physics: scrollEnabled
                    ? const ClampingScrollPhysics()
                    : const NeverScrollableScrollPhysics(),
                children: const [
                  SizedBox(height: 1000, child: Text('Scrollable content')),
                ],
              );
            },
          ),
        ),
      );
      await tester.pumpAndSettle();

      final panel = find.byKey(const Key('dashboardPanel'));
      final list = find.byKey(const Key('jitterList'));
      final expandedHeight = tester.getSize(panel).height;

      final drag = await tester.startGesture(tester.getCenter(list));
      await drag.moveBy(const Offset(0, 4));
      await tester.pump();
      expect(tester.getSize(panel).height, expandedHeight);

      await drag.moveBy(const Offset(0, -220));
      await tester.pump();
      await drag.up();
      await tester.pumpAndSettle();

      expect(contentController, isNotNull);
      expect(contentController!.offset, greaterThan(0));
      expect(tester.getSize(panel).height, expandedHeight);
      expect(
        tester.widget<ListView>(list).physics,
        isA<ClampingScrollPhysics>(),
      );
    },
  );

  testWidgets(
    'downward body pull hands off from scrolled content to panel at top',
    (tester) async {
      ScrollController? contentController;

      await tester.pumpWidget(
        MaterialApp(
          home: RideDashboardScaffold(
            panelIdentity: 'scroll-handoff',
            minPanelSize: 0.20,
            initialPanelSize: 0.60,
            maxPanelSize: 0.60,
            map: const ColoredBox(color: Colors.blueGrey),
            panelBuilder: (context, scrollController, scrollEnabled) {
              contentController = scrollController;
              return ListView(
                key: const Key('handoffList'),
                controller: scrollController,
                physics: scrollEnabled
                    ? const ClampingScrollPhysics()
                    : const NeverScrollableScrollPhysics(),
                children: const [
                  SizedBox(height: 1000, child: Text('Scrollable content')),
                ],
              );
            },
          ),
        ),
      );
      await tester.pumpAndSettle();

      final panel = find.byKey(const Key('dashboardPanel'));
      final list = find.byKey(const Key('handoffList'));
      final expandedHeight = tester.getSize(panel).height;

      await tester.drag(list, const Offset(0, -260));
      await tester.pumpAndSettle();
      expect(contentController, isNotNull);
      expect(contentController!.offset, greaterThan(0));

      final drag = await tester.startGesture(tester.getCenter(list));
      for (var i = 0; i < 14; i++) {
        await drag.moveBy(const Offset(0, 40));
        await tester.pump();
      }

      expect(contentController!.offset, closeTo(0, 0.5));
      expect(tester.getSize(panel).height, lessThan(expandedHeight));

      await drag.up();
      await tester.pumpAndSettle();

      expect(tester.getSize(panel).height, lessThan(expandedHeight));
      expect(contentController!.offset, closeTo(0, 0.5));
    },
  );
}
