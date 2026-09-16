import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';

import 'test_doubles.dart';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_controller.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_panels.dart';

class FlowFake implements RideFlowRepository {
  RideFlowPayload ride = {'id': 'ride', 'status': 'requested', 'trip': null};
  RideFlowPayload? trip;
  int reads = 0;
  int actions = 0;
  bool loseResponse = false;
  bool locationFails = false;
  bool selectable = true;
  RideFlowPayload? lastData;
  bool? lastPut;
  Completer<void>? blocked;
  Future<RideFlowPayload> _get(String path) async {
    reads++;
    if (blocked != null) {
      await blocked!.future;
    }
    if (path == '/v1/driver/trip') {
      if (trip == null) {
        throw const ApiException('missing', 'No trip', statusCode: 404);
      }
      return trip!;
    }
    if (path.endsWith('/offers')) {
      return {
        'offers': [
          {
            'driver_user_id': 'driver',
            'selectable': selectable,
            'status': 'pending',
            'fare': {'amount_minor': 10000, 'currency': 'PKR'},
            'updated_at': '2026-09-11T00:00:00Z',
          },
        ],
      };
    }
    if (path.endsWith('/driver-location')) {
      if (locationFails) {
        throw const ApiException('failed', 'Location failed', statusCode: 500);
      }
      throw const ApiException('missing', 'No location', statusCode: 404);
    }
    if (path.endsWith('/trips')) return {'trips': []};
    if (path.contains('/marketplace/')) return {'ride_requests': []};
    return ride;
  }

  Future<void> _act(
    String path, {
    RideFlowPayload? data,
    bool put = false,
  }) async {
    actions++;
    lastData = data;
    lastPut = put;
    ride = {
      'id': 'ride',
      'status': 'requested',
      'trip': {'status': 'assigned'},
    };
    if (loseResponse) {
      throw const ApiException('network_error', 'Response lost');
    }
  }

  @override
  Future<RideFlowPayload?> loadDriverTrip() async {
    try {
      return await _get('/v1/driver/trip');
    } on ApiException catch (_) {
      return null;
    }
  }

  @override
  Future<RideFlowPayload> loadDriverOpportunities() =>
      _get('/v1/driver/marketplace/ride-requests');
  @override
  Future<RideFlowPayload> loadDriverHistory() => _get('/v1/driver/trips');
  @override
  Future<RideFlowPayload> loadRiderRide(String id) =>
      _get('/v1/ride-requests/$id');
  @override
  Future<RideFlowPayload> loadRiderHistory() => _get('/v1/ride-requests');
  @override
  Future<RideFlowPayload> loadRiderOffers(String id) =>
      _get('/v1/ride-requests/$id/offers');
  @override
  Future<RideFlowPayload?> loadDriverLocation(String id) =>
      _get('/v1/ride-requests/$id/driver-location');
  @override
  Future<void> acceptFare(String id) =>
      _act('/v1/driver/ride-requests/$id/accept');
  @override
  Future<void> proposeFare(String id, int amount) => _act(
    '/v1/driver/ride-requests/$id/offer',
    data: {'amount_minor': amount},
    put: true,
  );
  @override
  Future<void> selectOffer(String id, String driver, String updatedAt) => _act(
    '/v1/ride-requests/$id/offers/$driver/accept',
    data: {'updated_at': updatedAt},
  );
  @override
  Future<void> rejectOffer(String id, String driver) =>
      _act('/v1/ride-requests/$id/offers/$driver/reject');
  @override
  Future<void> startTrip(String id) =>
      _act('/v1/driver/ride-requests/$id/start');
  @override
  Future<void> completeTrip(String id) =>
      _act('/v1/driver/ride-requests/$id/complete');
  @override
  Future<void> cancelTrip(String id) =>
      _act('/v1/driver/ride-requests/$id/cancel');
  @override
  Future<void> confirmCashCollected(String id) =>
      _act('/v1/driver/ride-requests/$id/cash-collected');
}

void main() {
  testWidgets(
    'Rider confirms exact displayed revision and unavailable offers cannot be selected',
    (tester) async {
      final repo = FlowFake();
      final flow = RideFlowController(
        repo,
        mode: RideFlowMode.rider,
        rideId: 'ride',
      );
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            rideFlowControllerProvider('ride').overrideWith((ref) => flow),
            riderRequestControllerProvider.overrideWith(
              (ref) => RiderRequestController(
                FakeRideRequestRepository(),
                FakeDeviceLocation(),
              ),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SingleChildScrollView(
                child: RideFlowPanel(mode: RideFlowMode.rider, rideId: 'ride'),
              ),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Choose Driver'));
      await tester.pumpAndSettle();
      expect(repo.actions, 0);
      await tester.tap(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.widgetWithText(FilledButton, 'Choose Driver'),
        ),
      );
      await tester.pumpAndSettle();
      expect(repo.actions, 1);
      expect(repo.lastData, {'updated_at': '2026-09-11T00:00:00Z'});
      repo.ride = {'id': 'ride', 'status': 'requested', 'trip': null};
      repo.selectable = false;
      await flow.refresh();
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Choose Driver'),
            )
            .onPressed,
        isNull,
      );
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );
  test('lost acceptance response reloads authoritative assignment without claiming success', () async {
    final repo = FlowFake()..loseResponse = true;
    final flow = RideFlowController(
      repo,
      mode: RideFlowMode.rider,
      rideId: 'ride',
    );
    await Future<void>.delayed(Duration.zero);
    expect(flow.offers, hasLength(1));
    await flow.runCommand(
      () => repo.selectOffer('ride', 'driver', '2026-09-11T00:00:00Z'),
    );
    expect(flow.status, 'assigned');
    expect(flow.offers, isEmpty);
    expect(flow.error, contains('Response lost'));
    expect(flow.location, isNull);
    flow.dispose();
  });
  test('location failure does not hide assigned Rider trip', () async {
    final repo = FlowFake()..locationFails = true;
    repo.ride = {
      'id': 'ride',
      'status': 'requested',
      'trip': {'status': 'assigned'},
    };
    final flow = RideFlowController(
      repo,
      mode: RideFlowMode.rider,
      rideId: 'ride',
    );
    await Future<void>.delayed(Duration.zero);
    expect(flow.status, 'assigned');
    expect(flow.loaded, isTrue);
    expect(flow.location, isNull);
    flow.dispose();
  });
  test(
    'restores assigned Driver trip even while availability is offline',
    () async {
      final repo = FlowFake()
        ..trip = {'ride_request_id': 'ride', 'status': 'in_progress'};
      final flow = RideFlowController(repo, mode: RideFlowMode.driver);
      await Future<void>.delayed(Duration.zero);
      expect(flow.status, 'in_progress');
      expect(flow.requests, isEmpty);
      flow.dispose();
    },
  );
  test('background stops polling; dispose tolerates in-flight reads', () async {
    final repo = FlowFake();
    final flow = RideFlowController(
      repo,
      mode: RideFlowMode.rider,
      rideId: 'ride',
      interval: const Duration(milliseconds: 10),
    );
    await Future<void>.delayed(Duration.zero);
    flow.setForeground(false);
    final reads = repo.reads;
    await Future<void>.delayed(const Duration(milliseconds: 30));
    expect(repo.reads, reads);
    repo.blocked = Completer<void>();
    flow.setForeground(true);
    flow.dispose();
    repo.blocked!.complete();
    await Future<void>.delayed(Duration.zero);
  });
  test(
    'counteroffer waits for an in-flight reload instead of being dropped',
    () async {
      final repo = FlowFake()..blocked = Completer<void>();
      final flow = RideFlowController(
        repo,
        mode: RideFlowMode.rider,
        rideId: 'ride',
      );
      await Future<void>.delayed(Duration.zero);
      expect(flow.busy, isTrue);
      final action = flow.runCommand(() => repo.proposeFare('ride', 11500));
      await Future<void>.delayed(Duration.zero);
      expect(repo.actions, 0);
      repo.blocked!.complete();
      await action;
      expect(repo.actions, 1);
      expect(repo.lastData, {'amount_minor': 11500});
      expect(repo.lastPut, isTrue);
      flow.dispose();
    },
  );
  test(
    'fare parsing is exact and rejects non-finite or excess precision input',
    () {
      expect(parseFareMinor('100.01'), 10001);
      for (final input in ['NaN', 'Infinity', '1.001', '-1', '0', '1e4']) {
        expect(parseFareMinor(input), isNull);
      }
    },
  );
  test('stale or future locations are never shown as live', () {
    expect(
      freshDriverLocation({
        'updated_at': DateTime.now()
            .subtract(const Duration(minutes: 3))
            .toIso8601String(),
      }),
      isNull,
    );
    expect(
      freshDriverLocation({
        'updated_at': DateTime.now()
            .add(const Duration(minutes: 3))
            .toIso8601String(),
      }),
      isNull,
    );
    expect(
      freshDriverLocation({'updated_at': DateTime.now().toIso8601String()}),
      isNotNull,
    );
  });
}
