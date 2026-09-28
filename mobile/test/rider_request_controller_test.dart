import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_execution.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  test(
    'uses device pickup and submits selected destination and fare',
    () async {
      final repository = FakeRideRequestRepository();
      final controller = RiderRequestController(
        repository,
        const FakeDeviceLocation(),
      );
      await controller.load();

      await controller.useCurrentPickup();
      controller.setDestination(
        const GeoPoint(latitude: 24.9056, longitude: 67.0822),
      );
      final created = await controller.submit(
        amountMinor: 70000,
        currency: 'PKR',
      );

      expect(created, isTrue);
      expect(
        repository.submittedPickup,
        const GeoPoint(latitude: 24.86, longitude: 67.01),
      );
      expect(
        repository.submittedDestination,
        const GeoPoint(latitude: 24.9056, longitude: 67.0822),
      );
      expect(
        repository.submittedFare,
        const Money(amountMinor: 70000, currency: 'PKR'),
      );
      expect(controller.state.active?.id, 'ride-1');
    },
  );

  test('requires pickup and destination before submission', () async {
    final controller = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    await controller.load();

    final created = await controller.submit(
      amountMinor: 70000,
      currency: 'PKR',
    );

    expect(created, isFalse);
    expect(controller.state.error, 'Choose both pickup and destination.');
  });

  group('active request lifecycle', () {
    test('keeps a requested request without a trip active', () {
      expect(stateWith(requestStatus: 'requested').active?.id, 'ride-1');
    });

    test('keeps an accepted request with an assigned trip active', () {
      expect(
        stateWith(requestStatus: 'accepted', tripStatus: 'assigned').active?.id,
        'ride-1',
      );
    });

    test('keeps an accepted request with an in-progress trip active', () {
      expect(
        stateWith(
          requestStatus: 'accepted',
          tripStatus: 'in_progress',
        ).active?.id,
        'ride-1',
      );
    });

    test('keeps a completed unsettled trip active until cash collection', () {
      expect(
        stateWith(
          requestStatus: 'accepted',
          tripStatus: 'completed',
          settlementStatus: 'unsettled',
        ).active?.id,
        'ride-1',
      );
    });

    test('treats a completed cash-collected trip as terminal', () {
      expect(
        stateWith(
          requestStatus: 'accepted',
          tripStatus: 'completed',
          settlementStatus: 'cash_collected',
        ).active,
        isNull,
      );
    });

    test('treats a cancelled request as terminal', () {
      expect(stateWith(requestStatus: 'cancelled').active, isNull);
    });

    test('treats an expired request as terminal', () {
      expect(stateWith(requestStatus: 'expired').active, isNull);
    });

    test('treats a cancelled trip as terminal', () {
      expect(
        stateWith(requestStatus: 'accepted', tripStatus: 'cancelled').active,
        isNull,
      );
    });
  });

  group('submission after settlement', () {
    test('allows a new ride after cash collection', () async {
      final previousRide = requestedRide.copyWith(
        status: 'accepted',
        trip: const TripSnapshot(
          status: 'completed',
          settlement: SettlementSnapshot(status: 'cash_collected'),
        ),
      );
      final repository = FakeRideRequestRepository(requests: [previousRide]);
      final controller = RiderRequestController(
        repository,
        const FakeDeviceLocation(),
      );
      await controller.load();
      await prepareSubmission(controller);

      expect(
        await controller.submit(amountMinor: 70000, currency: 'PKR'),
        isTrue,
      );
    });

    test('blocks a new ride while cash remains unsettled', () async {
      final previousRide = requestedRide.copyWith(
        status: 'accepted',
        trip: const TripSnapshot(
          status: 'completed',
          settlement: SettlementSnapshot(status: 'unsettled'),
        ),
      );
      final repository = FakeRideRequestRepository(requests: [previousRide]);
      final controller = RiderRequestController(
        repository,
        const FakeDeviceLocation(),
      );
      await controller.load();
      await prepareSubmission(controller);

      expect(
        await controller.submit(amountMinor: 70000, currency: 'PKR'),
        isFalse,
      );
    });
  });

  test('cancels the active request and returns to request creation', () async {
    final repository = FakeRideRequestRepository(requests: [requestedRide]);
    final controller = RiderRequestController(
      repository,
      const FakeDeviceLocation(),
    );
    await controller.load();

    await controller.cancelActive();

    expect(controller.state.active, isNull);
    expect(controller.state.requests.single.status, 'cancelled');
  });
}

RiderRequestState stateWith({
  required String requestStatus,
  String? tripStatus,
  String? settlementStatus,
}) {
  final trip = tripStatus == null
      ? null
      : TripSnapshot(
          status: tripStatus,
          settlement: settlementStatus == null
              ? null
              : SettlementSnapshot(status: settlementStatus),
        );
  return RiderRequestState(
    requests: [requestedRide.copyWith(status: requestStatus, trip: trip)],
  );
}

Future<void> prepareSubmission(RiderRequestController controller) async {
  await controller.useCurrentPickup();
  controller.setDestination(
    const GeoPoint(latitude: 24.9056, longitude: 67.0822),
  );
}
