import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_offer.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_snapshot.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_execution.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';
import 'package:uber_clone/features/ride_flow/application/driver_marketplace_controller.dart';
import 'package:uber_clone/features/ride_flow/application/driver_trip_controller.dart';
import 'package:uber_clone/features/ride_flow/application/rider_active_ride_controller.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_panels.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';
import 'package:uber_clone/features/driver_workspace/application/driver_controller.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';

import 'test_doubles.dart';

RideLocation _point(double lat, double lng) =>
    RideLocation(latitude: lat, longitude: lng);

RideFare _fare(int amount) => RideFare(amountMinor: amount, currency: 'PKR');

SettlementSnapshot _settlement([String status = 'unsettled']) =>
    SettlementSnapshot(status: status);

RiderRideSnapshot _ride({TripSnapshot? trip}) => RiderRideSnapshot(
  serviceCode: 'economy',
  id: 'ride',
  pickup: _point(33.68, 73.04),
  destination: _point(33.56, 73.01),
  proposedFare: _fare(10000),
  status: 'requested',
  createdAt: DateTime.utc(2026, 9, 11),
  expiresAt: DateTime.utc(2026, 9, 11, 0, 10),
  trip: trip,
);

TripSnapshot _trip(String status) => TripSnapshot(
  rideRequestId: 'ride',
  pickup: _point(33.68, 73.04),
  destination: _point(33.56, 73.01),
  status: status,
  assignedAt: DateTime.utc(2026, 9, 11),
  settlement: _settlement(),
);

RiderOfferComparison _offer(String driverUserId, {bool selectable = true}) =>
    RiderOfferComparison(
      rideRequestId: 'ride',
      driverUserId: driverUserId,
      fare: _fare(10000),
      status: 'pending',
      createdAt: DateTime.utc(2026, 9, 11),
      updatedAt: DateTime.utc(2026, 9, 11),
      expiresAt: DateTime.utc(2026, 9, 11, 0, 2),
      pickupDistanceMeters: 500,
      matchesProposedFare: true,
      selectable: selectable,
    );

MarketplaceRequest _marketplaceRequest(String id, {RideOffer? ownOffer}) =>
    MarketplaceRequest(
      id: id,
      pickup: _point(33.68, 73.04),
      destination: _point(33.56, 73.01),
      proposedFare: _fare(10000),
      createdAt: DateTime.utc(2026, 9, 11),
      expiresAt: DateTime.utc(2026, 9, 11, 0, 10),
      responseDeadline: DateTime.utc(2026, 9, 11, 0, 2),
      ownOffer: ownOffer,
      pickupDistanceMeters: 500,
    );

class FlowFake implements RideFlowRepository {
  RiderRideSnapshot ride = _ride();
  TripSnapshot? trip;
  int reads = 0;
  int actions = 0;
  bool loseResponse = false;
  bool locationFails = false;
  bool selectable = true;
  int? lastAmountMinor;
  DateTime? lastUpdatedAt;
  Completer<void>? blocked;
  List<MarketplaceRequest> marketplaceRequests = const [];
  List<RiderOfferComparison> riderOffers = const [];

  Future<void> _readBarrier() async {
    reads++;
    if (blocked != null) await blocked!.future;
  }

  @override
  Future<TripSnapshot?> getCurrentDriverTrip() async {
    await _readBarrier();
    return trip;
  }

  @override
  Future<List<MarketplaceRequest>> listMarketplaceRequests() async {
    await _readBarrier();
    return marketplaceRequests;
  }

  @override
  Future<List<DriverTripHistoryItem>> listDriverTrips() async {
    await _readBarrier();
    return const [];
  }

  @override
  Future<List<RiderRideSnapshot>> listRiderRides() async {
    await _readBarrier();
    return [ride];
  }

  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) async {
    await _readBarrier();
    return ride;
  }

  @override
  Future<List<RiderOfferComparison>> listRiderOffers(
    String rideRequestId,
  ) async {
    await _readBarrier();
    if (riderOffers.isNotEmpty) return riderOffers;
    return [_offer('driver', selectable: selectable)];
  }

  @override
  Future<DriverLocationSnapshot?> getDriverLocation(
    String rideRequestId,
  ) async {
    await _readBarrier();
    if (locationFails) {
      throw const ApiException('failed', 'Location failed', statusCode: 500);
    }
    return null;
  }

  Future<void> _action() async {
    actions++;
    ride = _ride(trip: _trip('assigned'));
    if (loseResponse) {
      throw const ApiException('network_error', 'Response lost');
    }
  }

  @override
  Future<void> selectOffer(
    String rideRequestId,
    String driverUserId,
    DateTime updatedAt,
  ) async {
    lastUpdatedAt = updatedAt;
    await _action();
  }

  @override
  Future<void> submitOffer(String rideRequestId, int amountMinor) async {
    lastAmountMinor = amountMinor;
    await _action();
  }

  @override
  Future<void> acceptProposedFare(String rideRequestId) => _action();

  @override
  Future<void> declineRideRequest(String rideRequestId) async {
    actions++;
    marketplaceRequests = marketplaceRequests
        .where((request) => request.id != rideRequestId)
        .toList();
  }

  @override
  Future<void> declineOffer(String rideRequestId, String driverUserId) async {
    actions++;
    riderOffers = riderOffers
        .where((offer) => offer.driverUserId != driverUserId)
        .toList();
  }

  @override
  Future<void> startTrip(String rideRequestId) => _action();

  @override
  Future<void> completeTrip(String rideRequestId) => _action();

  @override
  Future<void> confirmCashCollected(String rideRequestId) => _action();

  @override
  Future<void> cancelDriverTrip(String rideRequestId) => _action();
}

void main() {
  testWidgets(
    'Driver can decline only an open opportunity after confirmation',
    (tester) async {
      final offered = RideOffer(
        rideRequestId: 'offered-ride',
        driverUserId: 'driver',
        fare: _fare(10000),
        status: 'pending',
        createdAt: DateTime.utc(2026, 9, 11),
        updatedAt: DateTime.utc(2026, 9, 11),
        expiresAt: DateTime.utc(2026, 9, 11, 0, 2),
      );
      final repo = FlowFake()
        ..marketplaceRequests = [
          _marketplaceRequest('open-ride'),
          _marketplaceRequest('offered-ride', ownOffer: offered),
        ];
      final marketplace = DriverMarketplaceController(repo);
      final trips = DriverTripController(repo);
      final driver = DriverController(
        FakeDriverRepository(profile: driverProfile.copyWith(isOnline: true)),
        const FakeDeviceLocation(),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            driverMarketplaceControllerProvider.overrideWith(
              (ref) => marketplace,
            ),
            driverTripControllerProvider.overrideWith((ref) => trips),
            driverControllerProvider.overrideWith((ref) => driver),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: SingleChildScrollView(child: RideFlowPanel.driver()),
            ),
          ),
        ),
      );
      await tester.pump();
      await driver.setOnline(true);
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.widgetWithText(OutlinedButton, 'Decline'), findsOneWidget);
      await tester.tap(find.widgetWithText(OutlinedButton, 'Decline'));
      await tester.pump();
      expect(find.text('Decline request?'), findsOneWidget);
      expect(find.textContaining("only from your marketplace"), findsOneWidget);
      await tester.tap(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.widgetWithText(FilledButton, 'Decline'),
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(repo.actions, 1);
      expect(find.widgetWithText(OutlinedButton, 'Decline'), findsNothing);
      expect(find.textContaining('Your response:'), findsOneWidget);
      await tester.pumpWidget(const SizedBox.shrink());
    },
  );

  testWidgets(
    'Rider confirms exact displayed revision and unavailable offers cannot be selected',
    (tester) async {
      final repo = FlowFake();
      final flow = RiderActiveRideController(repo, rideId: 'ride');
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            riderActiveRideControllerProvider('ride')
                .overrideWith((ref) => flow),
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
                child: RideFlowPanel.rider(rideId: 'ride'),
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
      expect(repo.lastUpdatedAt, DateTime.utc(2026, 9, 11));

      repo.ride = _ride();
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

  testWidgets('Rider declines one offer and keeps another selectable', (
    tester,
  ) async {
    final repo = FlowFake()
      ..riderOffers = [_offer('driver-a'), _offer('driver-b')];
    final flow = RiderActiveRideController(repo, rideId: 'ride');
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          riderActiveRideControllerProvider('ride').overrideWith((ref) => flow),
          riderRequestControllerProvider.overrideWith(
            (ref) => RiderRequestController(
              FakeRideRequestRepository(),
              const FakeDeviceLocation(),
            ),
          ),
        ],
        child: const MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: RideFlowPanel.rider(rideId: 'ride'),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.widgetWithText(OutlinedButton, 'Decline'), findsNWidgets(2));
    await tester.tap(find.widgetWithText(OutlinedButton, 'Decline').first);
    await tester.pumpAndSettle();
    expect(find.text('Decline offer?'), findsOneWidget);
    expect(find.textContaining('Only this Driver offer'), findsOneWidget);
    await tester.tap(
      find.descendant(
        of: find.byType(AlertDialog),
        matching: find.widgetWithText(FilledButton, 'Decline'),
      ),
    );
    await tester.pumpAndSettle();

    expect(repo.actions, 1);
    expect(find.widgetWithText(FilledButton, 'Choose Driver'), findsOneWidget);
    expect(find.widgetWithText(OutlinedButton, 'Decline'), findsOneWidget);
    await tester.pumpWidget(const SizedBox.shrink());
  });

  test('lost acceptance response reloads authoritative assignment without claiming success', () async {
    final repo = FlowFake()..loseResponse = true;
    final flow = RiderActiveRideController(repo, rideId: 'ride');
    await Future<void>.delayed(Duration.zero);
    expect(flow.offers, hasLength(1));

    await flow.selectOffer('driver', DateTime.utc(2026, 9, 11));

    expect(flow.status, 'assigned');
    expect(flow.offers, isEmpty);
    expect(flow.error, contains('Response lost'));
    expect(flow.location, isNull);
    flow.dispose();
  });

  test('location failure does not hide assigned Rider trip', () async {
    final repo = FlowFake()..locationFails = true;
    repo.ride = _ride(trip: _trip('assigned'));
    final flow = RiderActiveRideController(repo, rideId: 'ride');
    await Future<void>.delayed(Duration.zero);
    expect(flow.status, 'assigned');
    expect(flow.loaded, isTrue);
    expect(flow.location, isNull);
    flow.dispose();
  });

  test(
    'restores assigned Driver trip even while availability is offline',
    () async {
      final repo = FlowFake()..trip = _trip('in_progress');
      final flow = DriverTripController(repo);
      await Future<void>.delayed(Duration.zero);
      expect(flow.trip?.status, 'in_progress');
      flow.dispose();
    },
  );

  test('background stops polling; dispose tolerates in-flight reads', () async {
    final repo = FlowFake();
    final flow = RiderActiveRideController(
      repo,
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
      final flow = DriverMarketplaceController(repo);

      await Future<void>.delayed(Duration.zero);
      expect(flow.busy, isTrue);

      final action = flow.submitOffer('ride', 11500);
      await Future<void>.delayed(Duration.zero);

      expect(repo.actions, 0);

      repo.blocked!.complete();
      await action;

      expect(repo.actions, 1);
      expect(repo.lastAmountMinor, 11500);
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
    DriverLocationSnapshot location(DateTime updatedAt) =>
        DriverLocationSnapshot(
          rideRequestId: 'ride',
          latitude: 33.68,
          longitude: 73.04,
          updatedAt: updatedAt,
        );

    expect(
      freshDriverLocation(
        location(DateTime.now().subtract(const Duration(minutes: 3))),
      ),
      isNull,
    );
    expect(
      freshDriverLocation(
        location(DateTime.now().add(const Duration(minutes: 3))),
      ),
      isNull,
    );
    expect(freshDriverLocation(location(DateTime.now())), isNotNull);
  });
}
