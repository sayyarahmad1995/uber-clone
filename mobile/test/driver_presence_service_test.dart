import 'dart:io';
import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/driver_workspace/data/driver_presence_publisher.dart';
import 'package:uber_clone/features/driver_workspace/data/android_driver_presence_service.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  test('Android declares a non-exported location foreground service', () {
    final manifest = File('android/app/src/main/AndroidManifest.xml')
        .readAsStringSync();

    expect(manifest, contains('android.permission.FOREGROUND_SERVICE'));
    expect(
      manifest,
      contains('android.permission.FOREGROUND_SERVICE_LOCATION'),
    );
    expect(manifest, contains('android.permission.POST_NOTIFICATIONS'));
    expect(
      manifest,
      matches(
        RegExp(
          r'<service\s+[^>]*android:name="com\.pravera\.flutter_foreground_task\.service\.ForegroundService"[^>]*android:foregroundServiceType="location"[^>]*android:exported="false"',
          dotAll: true,
        ),
      ),
    );
  });

  test(
    'fresh authenticated online Driver renews with a fresh GPS fix',
    () async {
      final repo = FakeDriverRepository(
        profile: driverProfile.copyWith(isOnline: true),
      );
      final location = CountingLocation();
      final publisher = DriverPresencePublisher(
        repository: repo,
        location: location,
        sessions: MemoryTokenStore(),
      );
      expect(await publisher.renew('user-1'), isTrue);
      expect(await publisher.renew('user-1'), isTrue);
      expect(location.reads, 2);
      expect(repo.calls, ['location', 'location']);
    },
  );

  test('expired session stops before profile or GPS lookup', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final location = CountingLocation();
    final publisher = DriverPresencePublisher(
      repository: repo,
      location: location,
      sessions: MemoryTokenStore()..token = null,
    );
    expect(await publisher.renew('user-1'), isFalse);
    expect(location.reads, 0);
    expect(repo.calls, isEmpty);
  });

  test('different authenticated Driver never publishes location', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(userId: 'user-2', isOnline: true),
    );
    final location = CountingLocation();
    final publisher = DriverPresencePublisher(
      repository: repo,
      location: location,
      sessions: MemoryTokenStore(),
    );
    expect(await publisher.renew('user-1'), isFalse);
    expect(location.reads, 0);
    expect(repo.calls, isEmpty);
  });

  test(
    'server-confirmed offline stops before GPS even if service lives',
    () async {
      final repo = FakeDriverRepository(profile: driverProfile);
      final location = CountingLocation();
      final publisher = DriverPresencePublisher(
        repository: repo,
        location: location,
        sessions: MemoryTokenStore(),
      );
      expect(await publisher.renew('user-1'), isFalse);
      expect(location.reads, 0);
      expect(repo.calls, isEmpty);
    },
  );

  test(
    'revoked location or failed network cannot counterfeit renewal',
    () async {
      final repo = FakeDriverRepository(
        profile: driverProfile.copyWith(isOnline: true),
      );
      final location = CountingLocation()
        ..failure = const LocationUnavailable('revoked');
      final publisher = DriverPresencePublisher(
        repository: repo,
        location: location,
        sessions: MemoryTokenStore(),
      );
      await expectLater(
        publisher.renew('user-1'),
        throwsA(isA<LocationUnavailable>()),
      );
      expect(repo.calls, isEmpty);
      location.failure = null;
      repo.failPublish = true;
      await expectLater(publisher.renew('user-1'), throwsException);
      expect(repo.calls, ['location']);
    },
  );

  test('authentication rejection stops rather than requesting GPS', () async {
    final repo = RejectingProfileRepository();
    final location = CountingLocation();
    final publisher = DriverPresencePublisher(
      repository: repo,
      location: location,
      sessions: MemoryTokenStore(),
    );
    expect(await publisher.renew('user-1'), isFalse);
    expect(location.reads, 0);
  });

  test('overlapping task ticks cannot publish twice', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    );
    final location = PendingPresenceLocation();
    final loop = DriverPresenceLoop(
      publisher: DriverPresencePublisher(
        repository: repo,
        location: location,
        sessions: MemoryTokenStore(),
      ),
      driverUserId: 'user-1',
      onStop: () async {},
    );
    final first = loop.tick();
    await Future<void>.delayed(Duration.zero);
    await loop.tick();
    location.fix.complete(const GeoPoint(latitude: 24.86, longitude: 67.01));
    await first;
    expect(repo.calls, ['location']);
  });

  test('offline task self-stops before GPS if Android stop failed', () async {
    final repo = FakeDriverRepository(profile: driverProfile);
    final location = CountingLocation();
    var stops = 0;
    final loop = DriverPresenceLoop(
      publisher: DriverPresencePublisher(
        repository: repo,
        location: location,
        sessions: MemoryTokenStore(),
      ),
      driverUserId: 'user-1',
      onStop: () async {
        stops++;
      },
    );
    await loop.tick();
    expect(stops, 1);
    expect(location.reads, 0);
    expect(repo.calls, isEmpty);
  });

  test(
    'service start requires task readiness for the intended account',
    () async {
      final bridge = MemoryPresenceBridge();
      final service = AndroidDriverPresenceService(bridge: bridge);
      await service.start('user-1');
      expect(await service.isRunningFor('user-1'), isTrue);
      await service.start('user-1');
      expect(bridge.starts, 1);
      expect(await service.isRunningFor('user-2'), isFalse);
    },
  );

  test('failed task readiness stops partial service', () async {
    final bridge = MemoryPresenceBridge()..acknowledgeStart = false;
    final service = AndroidDriverPresenceService(bridge: bridge);
    await expectLater(service.start('user-1'), throwsStateError);
    expect(bridge.running, isFalse);
  });

  test('failed first publish never acknowledges task readiness', () async {
    final repo = FakeDriverRepository(
      profile: driverProfile.copyWith(isOnline: true),
    )..failPublish = true;
    final loop = DriverPresenceLoop(
      publisher: DriverPresencePublisher(
        repository: repo,
        location: CountingLocation(),
        sessions: MemoryTokenStore(),
      ),
      driverUserId: 'user-1',
      onStop: () async {},
    );
    expect(await loop.tick(), isFalse);
    expect(repo.calls, ['location']);
  });
}

class MemoryTokenStore implements SessionStore {
  String? token = 'test-token';
  @override
  Future<String?> readValidToken() async => token;
  @override
  Future<void> clear() async => token = null;
  @override
  Future<void> save(String token, DateTime expiresAt) async =>
      this.token = token;
}

class CountingLocation implements DeviceLocation {
  int reads = 0;
  Object? failure;
  @override
  Future<GeoPoint> current() async {
    reads++;
    if (failure != null) throw failure!;
    return const GeoPoint(latitude: 24.86, longitude: 67.01);
  }
}

class RejectingProfileRepository extends FakeDriverRepository {
  @override
  Future<Never> get() async => throw const ApiException(
    'authentication_required',
    'Sign in',
    statusCode: 401,
  );
}

class PendingPresenceLocation implements DeviceLocation {
  final fix = Completer<GeoPoint>();
  @override
  Future<GeoPoint> current() => fix.future;
}

class MemoryPresenceBridge implements PresenceTaskBridge {
  bool running = false;
  bool acknowledgeStart = true;
  String? activeUserId;
  int starts = 0;
  @override
  Future<bool> isRunning() async => running;
  @override
  Future<String?> readyUserId() async => activeUserId;
  @override
  Future<void> start(String userId) async {
    starts++;
    running = true;
    if (acknowledgeStart) activeUserId = userId;
  }

  @override
  Future<void> stop() async {
    running = false;
    activeUserId = null;
  }
}
