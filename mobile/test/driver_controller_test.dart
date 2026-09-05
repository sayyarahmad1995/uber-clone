import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/driver_workspace/application/driver_controller.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  Future<DriverController> create(
    FakeDriverRepository repo, [
    DeviceLocation location = const FakeDeviceLocation(),
  ]) async {
    final controller = DriverController(repo, location);
    addTearDown(controller.dispose);
    await Future<void>.delayed(Duration.zero);
    return controller;
  }

  test('missing profile loads setup, onboarding persists profile', () async {
    final repo = FakeDriverRepository();
    final controller = await create(repo);
    expect(controller.loaded, isTrue);
    expect(controller.profile, isNull);
    await controller.onboard('Driver', driverVehicle);
    expect(controller.profile!.displayName, 'Driver');
    expect(controller.profile!.isOnline, isFalse);
  });
  test('going online publishes location before availability', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final controller = await create(repo);
    await controller.setOnline(true);
    expect(repo.calls, ['location', 'online=true']);
    expect(controller.profile!.isOnline, isTrue);
    expect(controller.location!.updatedAt, DateTime.utc(2026, 9, 5));
  });
  test(
    'permission failure prevents going online but never blocks offline',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final controller = await create(repo, DeniedLocation());
      await controller.setOnline(true);
      expect(repo.calls, isEmpty);
      expect(controller.profile!.isOnline, isFalse);
      expect(controller.error, contains('Permission denied'));
      await controller.setOnline(false);
      expect(repo.calls, ['online=false']);
      expect(controller.error, isNull);
    },
  );
  test(
    'publish and availability failures preserve confirmed online state',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile)
        ..failPublish = true;
      final controller = await create(repo);
      await controller.setOnline(true);
      expect(repo.calls, ['location']);
      expect(controller.profile!.isOnline, isFalse);
      repo.failPublish = false;
      repo.failAvailability = true;
      await controller.setOnline(true);
      expect(controller.profile!.isOnline, isFalse);
      expect(controller.error, isNotNull);
    },
  );
  test('duplicate commands are ignored while location is pending', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final pending = PendingLocation();
    final controller = await create(repo, pending);
    final first = controller.setOnline(true);
    await controller.setOnline(true);
    pending.result.complete(const GeoPoint(latitude: 24, longitude: 67));
    await first;
    expect(repo.calls, ['location', 'online=true']);
  });
  test(
    'leaving screen during device lookup prevents subsequent server writes',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final pending = PendingLocation();
      final controller = DriverController(repo, pending);
      await Future<void>.delayed(Duration.zero);
      final operation = controller.setOnline(true);
      controller.dispose();
      pending.result.complete(const GeoPoint(latitude: 24, longitude: 67));
      await operation;
      expect(repo.calls, isEmpty);
    },
  );
}

class DeniedLocation implements DeviceLocation {
  @override
  Future<GeoPoint> current() async =>
      throw const LocationUnavailable('Permission denied');
}

class PendingLocation implements DeviceLocation {
  final result = Completer<GeoPoint>();
  @override
  Future<GeoPoint> current() => result.future;
}
