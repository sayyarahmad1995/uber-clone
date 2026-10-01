import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/driver_workspace/application/driver_controller.dart';
import 'package:uber_clone/features/driver_workspace/domain/operating_state.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  Future<DriverController> create(
    FakeDriverRepository repo, [
    DeviceLocation location = const FakeDeviceLocation(),
    FakeDriverPresenceService? presence,
  ]) async {
    final controller = DriverController(
      repo,
      location,
      presence ?? FakeDriverPresenceService(),
    );
    addTearDown(controller.dispose);
    await Future<void>.delayed(Duration.zero);
    return controller;
  }

  test('first load preserves server-confirmed online profile', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final controller = await create(repo);
    expect(controller.profile!.isOnline, isTrue);
    expect(repo.calls, isNot(contains('online=false')));
    expect(controller.operation!.vehicleId, 'test-vehicle');
  });


  test('online restart hydrates map location with one foreground publish', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final presence = FakeDriverPresenceService()..runningFor = 'user-1';
    final controller = await create(
      repo,
      const FakeDeviceLocation(),
      presence,
    );

    for (var attempt = 0; attempt < 10 && controller.location == null; attempt++) {
      await Future<void>.delayed(Duration.zero);
    }

    expect(controller.profile!.isOnline, isTrue);
    expect(controller.onlinePresenceReady, isTrue);
    expect(controller.location, isNotNull);
    expect(repo.calls.where((call) => call == 'location'), hasLength(1));
    expect(presence.events, isEmpty);
  });

  test(
    'background leaves online state intact and resume reloads server truth',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final controller = await create(repo);
      await controller.setOnline(true);
      controller.setForeground(false);
      await Future<void>.delayed(Duration.zero);
      expect(repo.calls, isNot(contains('online=false')));
      expect(controller.profile!.isOnline, isTrue);
      controller.setForeground(true);
      await Future<void>.delayed(Duration.zero);
      expect(controller.profile!.isOnline, isTrue);
    },
  );

  test('background during location lookup cannot turn Driver online', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final pending = PendingLocation();
    final controller = await create(repo, pending);
    final operation = controller.setOnline(true);
    controller.setForeground(false);
    pending.result.complete(const GeoPoint(latitude: 24, longitude: 67));
    await operation;
    expect(repo.calls, isEmpty);
  });

  test('missing selection blocks location and online writes', () async {
    final repo = FakeDriverRepository(profile: driverProfile)
      ..operation = const OperatingState();
    final controller = await create(repo);
    await controller.setOnline(true);
    expect(repo.calls, isEmpty);
    expect(controller.error, contains('Select an approved'));
    await controller.selectOperation('owned');
    expect(controller.operation!.vehicleId, 'owned');
    await controller.load();
    expect(controller.operation!.vehicleId, 'owned');
  });

  test('going online publishes location before availability', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final presence = FakeDriverPresenceService(events: repo.calls);
    final controller = await create(repo, const FakeDeviceLocation(), presence);
    repo.calls.clear();
    await controller.setOnline(true);
    expect(repo.calls, ['location', 'online=true', 'service.start:user-1']);
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
      expect(repo.calls, ['location', 'location']);
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
    expect(repo.calls, ['location', 'location', 'online=true']);
  });

  test(
    'leaving screen during device lookup prevents subsequent server writes',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final pending = PendingLocation();
      final controller = DriverController(
        repo,
        pending,
        FakeDriverPresenceService(),
      );
      await Future<void>.delayed(Duration.zero);
      final operation = controller.setOnline(true);
      controller.dispose();
      pending.result.complete(const GeoPoint(latitude: 24, longitude: 67));
      await operation;
      expect(repo.calls, isEmpty);
    },
  );

  test('disposing an online controller never mutates availability', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final controller = DriverController(
      repo,
      const FakeDeviceLocation(),
      FakeDriverPresenceService(),
    );
    await Future<void>.delayed(Duration.zero);
    controller.dispose();
    await Future<void>.delayed(Duration.zero);
    expect(repo.calls, isNot(contains('online=false')));
  });

  test('repeat online and resume do not start a second service', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final presence = FakeDriverPresenceService();
    final controller = await create(repo, const FakeDeviceLocation(), presence);
    await controller.setOnline(true);
    await controller.setOnline(true);
    controller.setForeground(false);
    controller.setForeground(true);
    await Future<void>.delayed(Duration.zero);
    expect(presence.events, ['service.start:user-1']);
  });

  test('failed service start rolls confirmed online back offline', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final presence = FakeDriverPresenceService(events: repo.calls)
      ..failStart = true;
    final controller = await create(repo, const FakeDeviceLocation(), presence);
    repo.calls.clear();
    await controller.setOnline(true);
    expect(
      repo.calls,
      containsAllInOrder([
        'location',
        'online=true',
        'service.start:user-1',
        'online=false',
      ]),
    );
    expect(controller.profile!.isOnline, isFalse);
    expect(controller.error, contains('service unavailable'));
  });

  test(
    'explicit offline stops service only after server confirmation',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final presence = FakeDriverPresenceService(events: repo.calls);
      final controller = await create(
        repo,
        const FakeDeviceLocation(),
        presence,
      );
      await controller.setOnline(true);
      repo.calls.clear();
      await controller.setOnline(false);
      expect(repo.calls, ['online=false', 'service.stop']);
      expect(controller.profile!.isOnline, isFalse);
    },
  );

  test('failed offline request keeps online service running', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final presence = FakeDriverPresenceService();
    final controller = await create(repo, const FakeDeviceLocation(), presence);
    await controller.setOnline(true);
    repo.failAvailability = true;
    await controller.setOnline(false);
    expect(presence.runningFor, 'user-1');
    expect(presence.events, isNot(contains('service.stop')));
    expect(controller.profile!.isOnline, isTrue);
  });

  test(
    'failed service stop leaves server-confirmed offline and an error',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final presence = FakeDriverPresenceService()..failStop = true;
      final controller = await create(
        repo,
        const FakeDeviceLocation(),
        presence,
      );
      await controller.setOnline(true);
      await controller.setOnline(false);
      expect(repo.profile!.isOnline, isFalse);
      expect(controller.profile!.isOnline, isFalse);
      expect(controller.error, contains('service stop failed'));
      expect(presence.runningFor, 'user-1');
    },
  );

  test('failed rollback never claims durable online presence', () async {
    final repo = FakeDriverRepository(profile: driverProfile)
      ..failOffline = true;
    final presence = FakeDriverPresenceService()..failStart = true;
    final controller = await create(repo, const FakeDeviceLocation(), presence);
    await controller.setOnline(true);
    expect(repo.profile!.isOnline, isTrue);
    expect(controller.onlinePresenceReady, isFalse);
    expect(controller.error, contains('presence'));
  });
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
