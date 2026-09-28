import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/features/ride_flow/application/driver_marketplace_controller.dart';
import 'package:uber_clone/features/ride_flow/application/driver_trip_controller.dart';
import 'package:uber_clone/features/ride_flow/application/rider_active_ride_controller.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_offer.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_snapshot.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_execution.dart';
import 'package:uber_clone/features/ride_flow/domain/trip.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';

import 'test_doubles.dart';

RideLocation _point(double lat, double lng) =>
    RideLocation(latitude: lat, longitude: lng);

RideFare _fare(int amount) => RideFare(amountMinor: amount, currency: 'PKR');

SettlementSnapshot _settlement([String status = 'unsettled']) =>
    SettlementSnapshot(status: status);

TripSnapshot _trip(String status) => TripSnapshot(
  rideRequestId: 'ride',
  pickup: _point(33.68, 73.04),
  destination: _point(33.56, 73.01),
  status: status,
  assignedAt: DateTime.utc(2026, 9, 11),
  settlement: _settlement(),
);

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

class SplitFlowFake implements RideFlowRepository {
  RiderRideSnapshot ride = _ride();
  TripSnapshot? trip;

  int riderReads = 0;
  int marketplaceReads = 0;
  int tripReads = 0;
  int historyReads = 0;
  int actions = 0;

  int marketplaceReadsInFlight = 0;
  int maxMarketplaceReadsInFlight = 0;

  int? lastAmountMinor;
  DateTime? lastUpdatedAt;
  String? declinedRideRequestId;
  String? declinedDriverUserId;

  bool loseSelectResponse = false;
  bool failDecline = false;

  Completer<void>? riderBarrier;
  Completer<void>? marketplaceBarrier;
  Completer<void>? tripBarrier;

  @override
  Future<List<MarketplaceRequest>> listMarketplaceRequests() async {
    marketplaceReads++;
    marketplaceReadsInFlight++;
    if (marketplaceReadsInFlight > maxMarketplaceReadsInFlight) {
      maxMarketplaceReadsInFlight = marketplaceReadsInFlight;
    }

    try {
      final barrier = marketplaceBarrier;
      if (barrier != null) {
        await barrier.future;
      }
      return const [];
    } finally {
      marketplaceReadsInFlight--;
    }
  }

  @override
  Future<TripSnapshot?> getCurrentDriverTrip() async {
    tripReads++;
    final barrier = tripBarrier;
    if (barrier != null) {
      await barrier.future;
    }
    return trip;
  }

  @override
  Future<List<DriverTripHistoryItem>> listDriverTrips() async {
    historyReads++;
    return const [];
  }

  @override
  Future<List<RiderRideSnapshot>> listRiderRides() async => [ride];

  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) async {
    riderReads++;
    final barrier = riderBarrier;
    if (barrier != null) {
      await barrier.future;
    }
    return ride;
  }

  @override
  Future<List<RiderOfferComparison>> listRiderOffers(
    String rideRequestId,
  ) async {
    return [
      RiderOfferComparison(
        rideRequestId: rideRequestId,
        driverUserId: 'driver',
        fare: _fare(10000),
        status: 'pending',
        createdAt: DateTime.utc(2026, 9, 11),
        updatedAt: DateTime.utc(2026, 9, 11),
        expiresAt: DateTime.utc(2026, 9, 11, 0, 2),
        pickupDistanceMeters: 500,
        matchesProposedFare: true,
        selectable: true,
      ),
    ];
  }

  @override
  Future<DriverLocationSnapshot?> getDriverLocation(
    String rideRequestId,
  ) async {
    return null;
  }

  @override
  Future<void> submitOffer(String rideRequestId, int amountMinor) async {
    actions++;
    lastAmountMinor = amountMinor;
  }

  @override
  Future<void> acceptProposedFare(String rideRequestId) async {
    actions++;
  }

  @override
  Future<void> declineRideRequest(String rideRequestId) async {
    actions++;
    declinedRideRequestId = rideRequestId;
    if (failDecline) {
      throw const ApiException('conflict', 'Opportunity is not open');
    }
  }

  @override
  Future<void> declineOffer(String rideRequestId, String driverUserId) async {
    actions++;
    declinedRideRequestId = rideRequestId;
    declinedDriverUserId = driverUserId;
    if (failDecline) {
      throw const ApiException('conflict', 'Offer is not actionable');
    }
  }

  @override
  Future<void> selectOffer(
    String rideRequestId,
    String driverUserId,
    DateTime updatedAt,
  ) async {
    actions++;
    lastUpdatedAt = updatedAt;

    ride = _ride(trip: _trip('assigned'));

    if (loseSelectResponse) {
      throw const ApiException('network_error', 'Response lost');
    }
  }

  @override
  Future<void> startTrip(String rideRequestId) async {
    actions++;
  }

  @override
  Future<void> completeTrip(String rideRequestId) async {
    actions++;
  }

  @override
  Future<void> confirmCashCollected(String rideRequestId) async {
    actions++;
  }

  @override
  Future<void> cancelDriverTrip(String rideRequestId) async {
    actions++;
  }
}

void main() {
  test('RiderActiveRideController owns Rider active ride state', () {
    final controller = RiderActiveRideController(
      FakeRideFlowRepository(),
      rideId: 'ride',
    );

    expect(controller.rideId, 'ride');
    expect(controller.riderRide, isNull);
    expect(controller.offers, isEmpty);
    expect(controller.location, isNull);

    controller.dispose();
  });

  test('DriverMarketplaceController owns marketplace request state', () {
    final controller = DriverMarketplaceController(FakeRideFlowRepository());

    expect(controller.requests, isEmpty);

    controller.dispose();
  });

  test('DriverTripController owns active Trip and history state', () {
    final controller = DriverTripController(FakeRideFlowRepository());

    expect(controller.trip, isNull);
    expect(controller.history, isEmpty);

    controller.dispose();
  });

  test('refreshes never overlap', () async {
    final repo = SplitFlowFake()..marketplaceBarrier = Completer<void>();

    final controller = DriverMarketplaceController(
      repo,
      interval: const Duration(milliseconds: 10),
    );

    await Future<void>.delayed(Duration.zero);
    expect(repo.marketplaceReadsInFlight, 1);

    final refresh = controller.refresh();
    await Future<void>.delayed(Duration.zero);

    expect(repo.marketplaceReads, 1);
    expect(repo.maxMarketplaceReadsInFlight, 1);

    repo.marketplaceBarrier!.complete();
    await refresh;

    expect(repo.maxMarketplaceReadsInFlight, 1);

    controller.dispose();
  });

  test('marketplace command waits for an in-flight refresh', () async {
    final repo = SplitFlowFake()..marketplaceBarrier = Completer<void>();

    final controller = DriverMarketplaceController(repo);

    await Future<void>.delayed(Duration.zero);
    expect(controller.busy, isTrue);

    final command = controller.submitOffer('ride', 11500);
    await Future<void>.delayed(Duration.zero);

    expect(repo.actions, 0);

    repo.marketplaceBarrier!.complete();
    repo.marketplaceBarrier = null;

    await command;

    expect(repo.actions, 1);
    expect(repo.lastAmountMinor, 11500);

    controller.dispose();
  });

  test('Driver decline is serialized and reloads marketplace state', () async {
    final repo = SplitFlowFake();
    final controller = DriverMarketplaceController(repo);
    await Future<void>.delayed(Duration.zero);
    final readsBefore = repo.marketplaceReads;

    await controller.declineRideRequest('ride');

    expect(repo.actions, 1);
    expect(repo.declinedRideRequestId, 'ride');
    expect(repo.marketplaceReads, readsBefore + 1);
    controller.dispose();
  });

  test('Rider decline rejects one offer and reloads Rider state', () async {
    final repo = SplitFlowFake();
    final controller = RiderActiveRideController(repo, rideId: 'ride');
    await Future<void>.delayed(Duration.zero);
    final readsBefore = repo.riderReads;

    await controller.declineOffer('driver');

    expect(repo.actions, 1);
    expect(repo.declinedRideRequestId, 'ride');
    expect(repo.declinedDriverUserId, 'driver');
    expect(repo.riderReads, readsBefore + 1);
    controller.dispose();
  });

  test('failed Driver decline reloads and exposes the server error', () async {
    final repo = SplitFlowFake()..failDecline = true;
    final controller = DriverMarketplaceController(repo);
    await Future<void>.delayed(Duration.zero);
    final readsBefore = repo.marketplaceReads;

    await controller.declineRideRequest('ride');

    expect(controller.error, contains('Opportunity is not open'));
    expect(repo.marketplaceReads, readsBefore + 1);
    controller.dispose();
  });

  test('failed Rider decline reloads and exposes the server error', () async {
    final repo = SplitFlowFake()..failDecline = true;
    final controller = RiderActiveRideController(repo, rideId: 'ride');
    await Future<void>.delayed(Duration.zero);
    final readsBefore = repo.riderReads;

    await controller.declineOffer('driver');

    expect(controller.error, contains('Offer is not actionable'));
    expect(controller.offers, hasLength(1));
    expect(repo.riderReads, readsBefore + 1);
    controller.dispose();
  });

  test('background marketplace command preserves the existing error', () async {
    final controller = DriverMarketplaceController(FakeRideFlowRepository());
    await Future<void>.delayed(Duration.zero);
    controller.setForeground(false);
    controller.error = 'previous error';
    var notifications = 0;
    controller.addListener(() => notifications++);

    await controller.submitOffer('ride', 11500);

    expect(controller.error, 'previous error');
    expect(notifications, 0);
    controller.dispose();
  });

  test('background Rider command preserves the existing error', () async {
    final controller = RiderActiveRideController(
      FakeRideFlowRepository(),
      rideId: 'ride',
    );
    await Future<void>.delayed(Duration.zero);
    controller.setForeground(false);
    controller.error = 'previous error';
    var notifications = 0;
    controller.addListener(() => notifications++);

    await controller.selectOffer('driver', DateTime.utc(2026, 9, 11));

    expect(controller.error, 'previous error');
    expect(notifications, 0);
    controller.dispose();
  });

  test('background Driver Trip command preserves the existing error', () async {
    final controller = DriverTripController(FakeRideFlowRepository());
    await Future<void>.delayed(Duration.zero);
    controller.setForeground(false);
    controller.error = 'previous error';
    var notifications = 0;
    controller.addListener(() => notifications++);

    await controller.startTrip('ride');

    expect(controller.error, 'previous error');
    expect(notifications, 0);
    controller.dispose();
  });

  test(
    'disposed command preserves the existing error without notifying',
    () async {
      final controller = DriverMarketplaceController(FakeRideFlowRepository());
      await Future<void>.delayed(Duration.zero);
      controller.error = 'previous error';
      var notifications = 0;
      controller.addListener(() => notifications++);
      controller.dispose();

      await controller.acceptProposedFare('ride');

      expect(controller.error, 'previous error');
      expect(notifications, 0);
    },
  );

  test('ambiguous Rider command failure reloads authoritative state', () async {
    final repo = SplitFlowFake()..loseSelectResponse = true;
    final controller = RiderActiveRideController(repo, rideId: 'ride');

    await Future<void>.delayed(Duration.zero);
    expect(controller.offers, hasLength(1));

    await controller.selectOffer('driver', DateTime.utc(2026, 9, 11));

    expect(controller.status, 'assigned');
    expect(controller.offers, isEmpty);
    expect(controller.error, contains('Response lost'));
    expect(repo.lastUpdatedAt, DateTime.utc(2026, 9, 11));

    controller.dispose();
  });

  test('background pauses polling', () async {
    final repo = SplitFlowFake();
    final controller = DriverMarketplaceController(
      repo,
      interval: const Duration(milliseconds: 10),
    );

    await Future<void>.delayed(Duration.zero);
    controller.setForeground(false);

    final reads = repo.marketplaceReads;

    await Future<void>.delayed(const Duration(milliseconds: 30));

    expect(repo.marketplaceReads, reads);

    controller.dispose();
  });

  test('foreground triggers an immediate refresh', () async {
    final repo = SplitFlowFake();
    final controller = DriverMarketplaceController(
      repo,
      interval: const Duration(seconds: 30),
    );

    await Future<void>.delayed(Duration.zero);
    controller.setForeground(false);

    final reads = repo.marketplaceReads;

    controller.setForeground(true);
    await Future<void>.delayed(Duration.zero);

    expect(repo.marketplaceReads, reads + 1);

    controller.dispose();
  });

  test('Driver Trip controller restores active Trip and history', () async {
    final repo = SplitFlowFake()..trip = _trip('in_progress');
    final controller = DriverTripController(repo);

    await Future<void>.delayed(Duration.zero);

    expect(controller.status, 'in_progress');
    expect(controller.trip, isNotNull);
    expect(repo.tripReads, 1);
    expect(repo.historyReads, 1);

    controller.dispose();
  });
}
