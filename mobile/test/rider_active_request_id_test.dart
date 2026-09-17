import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/ride_flow/domain/marketplace_request.dart';
import 'package:uber_clone/features/ride_flow/domain/ride_snapshot.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';
import 'package:uber_clone/features/rider_request/presentation/rider_request_screen.dart';

import 'test_doubles.dart';

class RecordingRideFlowRepository extends FakeRideFlowRepository {
  final requestedRideIds = <String>[];

  @override
  Future<RiderRideSnapshot> getRiderRide(String rideRequestId) async {
    requestedRideIds.add(rideRequestId);
    return RiderRideSnapshot(
      serviceCode: 'economy',
      id: rideRequestId,
      pickup: const RideLocation(latitude: 33.65041, longitude: 72.97183),
      destination: const RideLocation(latitude: 33.64566, longitude: 72.96401),
      proposedFare: const RideFare(amountMinor: 20000, currency: 'PKR'),
      status: 'requested',
      createdAt: DateTime.utc(2026, 9, 17),
      expiresAt: DateTime.utc(2026, 9, 17, 9, 10),
    );
  }
}

void main() {
  testWidgets('active Rider flow uses the actual ride request id', (tester) async {
    const rideId = '11111111-1111-4111-8111-111111111111';
    final request = RideRequest(
      id: rideId,
      pickup: const GeoPoint(latitude: 33.65041, longitude: 72.97183),
      destination: const GeoPoint(latitude: 33.64566, longitude: 72.96401),
      proposedFare: const Money(amountMinor: 20000, currency: 'PKR'),
      status: 'requested',
      createdAt: DateTime.utc(2026, 9, 17),
    );
    final requestRepository = FakeRideRequestRepository(requests: [request]);
    final requestController = RiderRequestController(
      requestRepository,
      const FakeDeviceLocation(),
    );
    final flowRepository = RecordingRideFlowRepository();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          riderRequestControllerProvider.overrideWith((ref) => requestController),
          rideFlowRepositoryProvider.overrideWithValue(flowRepository),
        ],
        child: const MaterialApp(home: RiderRequestScreen()),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(flowRepository.requestedRideIds, isNotEmpty);
    expect(flowRepository.requestedRideIds, everyElement(rideId));

    await tester.pumpWidget(const SizedBox.shrink());
    requestController.dispose();
  });
}
