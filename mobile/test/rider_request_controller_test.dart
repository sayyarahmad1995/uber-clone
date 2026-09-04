import 'package:flutter_test/flutter_test.dart';
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
