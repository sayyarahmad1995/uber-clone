import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/app.dart';
import 'package:uber_clone/core/dashboard/ride_dashboard_scaffold.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/data/ride_request_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  testWidgets('signed-out startup routes to application login', (tester) async {
    await tester.pumpWidget(testApp(FakeAuthRepository()));
    await tester.pumpAndSettle();
    expect(find.text('Welcome back'), findsOneWidget);
    expect(find.text('Sign in'), findsOneWidget);
  });

  testWidgets('restored account enters Rider by default', (tester) async {
    await tester.pumpWidget(testApp(FakeAuthRepository(account: riderAccount)));
    await tester.pumpAndSettle();
    expect(find.text('Looking for Driver offers'), findsOneWidget);
    expect(find.text('Rider'), findsOneWidget);
  });

  testWidgets('dual-capability account can switch to Driver', (tester) async {
    await tester.pumpWidget(
      testApp(FakeAuthRepository(account: bothCapabilities)),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Driver'));
    await tester.pumpAndSettle();
    expect(find.text('Driver dashboard'), findsOneWidget);
  });

  testWidgets('Rider dashboard focuses on device location at startup', (
    tester,
  ) async {
    final location = CountingDeviceLocation();
    await tester.pumpWidget(
      testApp(
        FakeAuthRepository(account: riderAccount),
        rideRequests: FakeRideRequestRepository(),
        deviceLocation: location,
      ),
    );
    await tester.pumpAndSettle();

    expect(location.requestCount, 1);
    final panel = find.byKey(const Key('dashboardPanel'));
    final dashboardHeight = tester
        .getSize(find.byType(RideDashboardScaffold))
        .height;
    final collapsedHeight = tester.getSize(panel).height;
    expect(collapsedHeight, moreOrLessEquals(dashboardHeight * 0.18));
    final mapControl = find.byTooltip('Center map on your location');
    final collapsedControlY = tester.getCenter(mapControl).dy;

    await tester.drag(find.text('Where are you going?'), const Offset(0, 80));
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, collapsedHeight);

    final shortGesture = await tester.startGesture(
      tester.getCenter(find.text('Where are you going?')),
    );
    await shortGesture.moveBy(const Offset(0, -30));
    await tester.pump();
    expect(tester.getSize(panel).height, greaterThan(collapsedHeight));
    await shortGesture.up();
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, collapsedHeight);

    final maximizeGesture = await tester.startGesture(
      tester.getCenter(find.text('Where are you going?')),
    );
    await maximizeGesture.moveBy(const Offset(0, -100));
    await tester.pump();
    final expandingHeight = tester.getSize(panel).height;
    expect(expandingHeight, greaterThan(collapsedHeight));
    expect(expandingHeight, lessThan(dashboardHeight * 0.60));
    expect(tester.getCenter(mapControl).dy, lessThan(collapsedControlY));

    await maximizeGesture.moveBy(const Offset(0, -1000));
    await tester.pump();
    expect(
      tester.getSize(panel).height,
      moreOrLessEquals(dashboardHeight * 0.60),
    );
    expect(
      tester
          .widget<ListView>(
            find.descendant(of: panel, matching: find.byType(ListView)),
          )
          .physics,
      isA<NeverScrollableScrollPhysics>(),
    );

    await maximizeGesture.up();
    await tester.pumpAndSettle();
    final expandedHeight = tester.getSize(panel).height;
    expect(expandedHeight, moreOrLessEquals(dashboardHeight * 0.60));

    final panelList = find.descendant(
      of: panel,
      matching: find.byType(ListView),
    );
    expect(
      tester.widget<ListView>(panelList).physics,
      isA<ClampingScrollPhysics>(),
    );
    await tester.drag(panelList, const Offset(0, -300));
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, expandedHeight);

    await tester.drag(panelList, const Offset(0, 600));
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, expandedHeight);

    final minimizeGesture = await tester.startGesture(
      tester.getCenter(find.text('Where are you going?')),
    );
    await minimizeGesture.moveBy(const Offset(0, 100));
    await tester.pump();
    final minimizingHeight = tester.getSize(panel).height;
    expect(minimizingHeight, lessThan(expandedHeight));
    expect(minimizingHeight, greaterThan(collapsedHeight));

    await minimizeGesture.up();
    await tester.pumpAndSettle();
    expect(tester.getSize(panel).height, collapsedHeight);

    await tester.tap(find.byTooltip('Center map on your location'));
    await tester.pumpAndSettle();

    expect(location.requestCount, 2);
  });

  testWidgets('dashboard resets scrolling when panel identity changes', (
    tester,
  ) async {
    final dashboardKey = GlobalKey<_DashboardIdentityHarnessState>();
    await tester.pumpWidget(
      MaterialApp(home: _DashboardIdentityHarness(key: dashboardKey)),
    );

    final list = find.byKey(const Key('identityPanelList'));
    await tester.drag(list, const Offset(0, -500));
    await tester.pumpAndSettle();
    expect(dashboardKey.currentState!.scrollController.offset, greaterThan(0));

    dashboardKey.currentState!.showSecondPanel();
    await tester.pump();

    expect(dashboardKey.currentState!.scrollController.offset, 0);
    expect(find.text('Panel two'), findsOneWidget);
    await tester.pumpAndSettle();
    expect(
      tester.getSize(find.byKey(const Key('dashboardPanel'))).height,
      tester.getSize(find.byType(RideDashboardScaffold)).height * 0.20,
    );
    expect(
      tester.widget<ListView>(list).physics,
      isA<NeverScrollableScrollPhysics>(),
    );
  });

  testWidgets('controls do not resize panel and retain taps and text input', (
    tester,
  ) async {
    await tester.pumpWidget(
      const MaterialApp(home: _DashboardIdentityHarness()),
    );
    final panel = find.byKey(const Key('dashboardPanel'));
    final height = tester.getSize(panel).height;
    for (final key in ['panelButton', 'panelField']) {
      final gesture = await tester.startGesture(
        tester.getCenter(find.byKey(Key(key))),
      );
      await gesture.moveBy(const Offset(0, 30));
      await tester.pump();
      expect(tester.getSize(panel).height, height);
      await gesture.up();
      await tester.pumpAndSettle();
      expect(tester.getSize(panel).height, height);
    }
    await tester.tap(find.byKey(const Key('panelButton')));
    await tester.pumpAndSettle();
    expect(find.text('Tapped'), findsOneWidget);
    await tester.enterText(find.byKey(const Key('panelField')), '42');
    expect(find.text('42'), findsOneWidget);
  });

  testWidgets(
    'body drag owns one pointer and locks scrolling until snap finishes',
    (tester) async {
      final key = GlobalKey<_DashboardIdentityHarnessState>();
      await tester.pumpWidget(
        MaterialApp(home: _DashboardIdentityHarness(key: key)),
      );
      key.currentState!.showSecondPanel();
      await tester.pumpAndSettle();
      final panel = find.byKey(const Key('dashboardPanel'));
      final list = find.byKey(const Key('identityPanelList'));
      final first = await tester.startGesture(
        tester.getCenter(find.text('Panel two')),
        pointer: 1,
      );
      await first.moveBy(const Offset(0, -100));
      await tester.pump();
      final preview = tester.getSize(panel).height;
      final second = await tester.startGesture(
        tester.getCenter(find.text('Panel two')),
        pointer: 2,
      );
      await second.moveBy(const Offset(0, 70));
      await second.up();
      await tester.pump();
      expect(tester.getSize(panel).height, preview);
      await first.up();
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));
      expect(
        tester.widget<ListView>(list).physics,
        isA<NeverScrollableScrollPhysics>(),
      );
      await tester.pumpAndSettle();
      expect(
        tester.widget<ListView>(list).physics,
        isA<ClampingScrollPhysics>(),
      );
    },
  );
}

class _DashboardIdentityHarness extends StatefulWidget {
  const _DashboardIdentityHarness({super.key});

  @override
  State<_DashboardIdentityHarness> createState() =>
      _DashboardIdentityHarnessState();
}

class _DashboardIdentityHarnessState extends State<_DashboardIdentityHarness> {
  String _identity = 'one';
  late ScrollController scrollController;
  bool tapped = false;

  void showSecondPanel() => setState(() => _identity = 'two');

  @override
  Widget build(BuildContext context) {
    return RideDashboardScaffold(
      panelIdentity: _identity,
      minPanelSize: 0.20,
      initialPanelSize: 0.60,
      maxPanelSize: 0.60,
      map: const ColoredBox(color: Colors.blueGrey),
      panelBuilder: (context, controller, scrollEnabled) {
        scrollController = controller;
        return ListView.builder(
          key: const Key('identityPanelList'),
          controller: controller,
          physics: scrollEnabled
              ? const ClampingScrollPhysics()
              : const NeverScrollableScrollPhysics(),
          itemCount: 40,
          itemBuilder: (context, index) => index == 1
              ? DashboardPanelControl(
                  child: TextButton(
                    key: const Key('panelButton'),
                    onPressed: () => setState(() => tapped = true),
                    child: Text(tapped ? 'Tapped' : 'Control'),
                  ),
                )
              : index == 2
              ? const DashboardPanelControl(
                  child: TextField(key: Key('panelField')),
                )
              : ListTile(
                  title: Text(index == 0 ? 'Panel $_identity' : 'Item $index'),
                ),
        );
      },
    );
  }
}

Widget testApp(
  AuthRepository repository, {
  RideRequestRepository? rideRequests,
  DeviceLocation? deviceLocation,
}) => ProviderScope(
  overrides: [
    authRepositoryProvider.overrideWithValue(repository),
    capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
    rideRequestRepositoryProvider.overrideWithValue(
      rideRequests ?? FakeRideRequestRepository(requests: [requestedRide]),
    ),
    deviceLocationProvider.overrideWithValue(
      deviceLocation ?? const FakeDeviceLocation(),
    ),
  ],
  child: const UberCloneApp(),
);

class CountingDeviceLocation implements DeviceLocation {
  int requestCount = 0;

  @override
  Future<GeoPoint> current() async {
    requestCount++;
    return const GeoPoint(latitude: 24.86, longitude: 67.01);
  }
}
