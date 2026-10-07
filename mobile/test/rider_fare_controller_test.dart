import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uber_clone/core/providers.dart';
import 'package:uber_clone/features/rider_request/data/route_preview_repository.dart';

import 'test_doubles.dart';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/rider_request/application/rider_fare_controller.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';
import 'package:uber_clone/features/rider_request/domain/route_preview.dart';
import 'package:uber_clone/features/rider_request/domain/ride_service.dart';

void main() {
  RoutePreview preview(int amount, String version) => RoutePreview(
    distanceMeters: 2000,
    durationSeconds: 300,
    encodedPolyline: '??_ibE_ibE',
    suggestedFare: Money(amountMinor: amount, currency: 'PKR'),
    pricingPolicyVersion: version,
  );
  test(
    'suggestion prefills an untouched proposal and preserves edits on refresh',
    () {
      final c = RiderFareController();
      c.invalidate('A');
      c.applyPreview('A', preview(12500, 'v1'));
      expect(c.text, '125.00');
      expect(c.version, 'v1');
      expect(c.userEdited, false);
      c.edit('1100');
      c.applyPreview('A', preview(22500, 'v2'));
      expect(c.text, '1100');
      expect(c.version, 'v2');
      expect(c.userEdited, true);
      c.dispose();
    },
  );
  test('changed input clears ownership and ignores an older preview', () {
    final c = RiderFareController();
    c.invalidate('A');
    c.applyPreview('A', preview(12500, 'v1'));
    c.edit('1100');
    c.invalidate('B');
    expect(c.text, isEmpty);
    expect(c.version, isNull);
    expect(c.userEdited, false);
    c.applyPreview('A', preview(99900, 'old'));
    expect(c.text, isEmpty);
    c.applyPreview('B', preview(22500, 'v2'));
    expect(c.text, '225.00');
    c.dispose();
  });
  test('route-only Driver responses remain valid and incomplete pricing is rejected', () {
    final wire = {
      'route': {
        'distance_meters': 2000,
        'duration_seconds': 300,
        'encoded_polyline': '??_ibE_ibE',
      },
    };
    expect(RoutePreview.fromJson(wire).suggestedFare, isNull);
    expect(
      () => RoutePreview.fromJson({
        ...wire,
        'suggested_fare': {'amount_minor': 12500, 'currency': 'PKR'},
      }),
      throwsFormatException,
    );
    expect(
      () => RoutePreview.fromJson({
        ...wire,
        'suggested_fare': {'amount_minor': 12500, 'currency': 'USD'},
        'pricing_policy_version': 'v1',
      }),
      throwsFormatException,
    );
  });
  test('catalog prices require an explicit server signal', () {
    final wire = {
      'code': 'synthetic',
      'display_name': 'Third',
      'display_order': 1,
      'presentation_token': 'future',
    };
    expect(RideService.fromJson(wire).pricingRequired, false);
    expect(
      RideService.fromJson({...wire, 'pricing_required': true}).pricingRequired,
      true,
    );
  });
  test(
    'account logout resets proposal ownership and ignores disposed results',
    () async {
      final container = ProviderContainer(
        overrides: [
          authRepositoryProvider.overrideWithValue(
            FakeAuthRepository(account: riderAccount),
          ),
          capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
          driverPresenceServiceProvider.overrideWithValue(
            FakeDriverPresenceService(),
          ),
          driverRepositoryProvider.overrideWithValue(FakeDriverRepository()),
          rideFlowRepositoryProvider.overrideWithValue(
            FakeRideFlowRepository(),
          ),
          rideRequestRepositoryProvider.overrideWithValue(
            FakeRideRequestRepository(),
          ),
          deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
          riderServiceRepositoryProvider.overrideWithValue(
            FakeRideServiceRepository(),
          ),
          routePreviewRepositoryProvider.overrideWithValue(_FareRoutes()),
        ],
      );
      final subscription = container.listen(
        riderFareControllerProvider,
        (_, __) {},
      );
      await Future<void>.delayed(Duration.zero);
      final previous = container.read(riderFareControllerProvider);
      previous.edit('1100');
      await container.read(sessionControllerProvider).logout();
      await Future<void>.delayed(Duration.zero);
      final current = container.read(riderFareControllerProvider);
      expect(identical(previous, current), false);
      expect(current.text, isEmpty);
      previous.applyPreview(previous.selectionKey ?? '', preview(99900, 'old'));
      expect(current.version, isNull);
      subscription.close();
      container.dispose();
    },
  );
  test(
    'catalog removal invalidates fare even while the picker is absent',
    () async {
      final catalog = FakeRideServiceRepository();
      final container = ProviderContainer(
        overrides: [
          authRepositoryProvider.overrideWithValue(FakeAuthRepository()),
          capabilityStoreProvider.overrideWithValue(MemoryCapabilityStore()),
          rideRequestRepositoryProvider.overrideWithValue(
            FakeRideRequestRepository(),
          ),
          deviceLocationProvider.overrideWithValue(const FakeDeviceLocation()),
          riderServiceRepositoryProvider.overrideWithValue(catalog),
          routePreviewRepositoryProvider.overrideWithValue(_FareRoutes()),
        ],
      );
      final subscription = container.listen(
        riderFareControllerProvider,
        (_, __) {},
      );
      await container.read(riderServicesProvider.future);
      final fare = container.read(riderFareControllerProvider);
      fare.edit('1100');
      catalog.services = [];
      container.invalidate(riderServicesProvider);
      await container.read(riderServicesProvider.future);
      expect(
        container.read(riderRequestControllerProvider).serviceCode,
        isEmpty,
      );
      expect(container.read(riderFareControllerProvider).text, isEmpty);
      expect(container.read(riderFareControllerProvider).version, isNull);
      subscription.close();
      container.dispose();
    },
  );
}

class _FareRoutes implements RoutePreviewRepository {
  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  }) async => RoutePreview(
    distanceMeters: 2000,
    durationSeconds: 300,
    encodedPolyline: '??_ibE_ibE',
    suggestedFare: const Money(amountMinor: 12500, currency: 'PKR'),
    pricingPolicyVersion: 'v1',
  );
  @override
  void cancel() {}
}
