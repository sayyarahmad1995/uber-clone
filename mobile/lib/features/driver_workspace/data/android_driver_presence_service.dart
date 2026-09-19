import 'dart:async';
import 'dart:io';
import 'dart:ui';

import 'package:dio/dio.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_foreground_task/flutter_foreground_task.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../../core/config/app_config.dart';
import '../../../core/session/session_store.dart';
import '../../rider_request/data/device_location.dart';
import '../application/driver_presence_service.dart';
import 'driver_presence_publisher.dart';
import 'driver_repository.dart';

const _expectedUserKey = 'driver_presence_expected_user';
const _readyUserKey = 'driver_presence_ready_user';

/// Android platform boundary; injectable to test startup/rollback without channels.
abstract interface class PresenceTaskBridge {
  Future<bool> isRunning();
  Future<String?> readyUserId();
  Future<void> start(String userId);
  Future<void> stop();
}

class AndroidDriverPresenceService implements DriverPresenceService {
  AndroidDriverPresenceService({PresenceTaskBridge? bridge})
    : _bridge = bridge ?? FlutterPresenceTaskBridge();

  final PresenceTaskBridge _bridge;

  @override
  Future<bool> isRunningFor(String driverUserId) async =>
      await _bridge.isRunning() && await _bridge.readyUserId() == driverUserId;

  @override
  Future<void> start(String driverUserId) async {
    if (await isRunningFor(driverUserId)) return;
    if (await _bridge.isRunning()) await _bridge.stop();
    try {
      await _bridge.start(driverUserId);
      if (!await isRunningFor(driverUserId)) {
        throw StateError('Driver presence task did not become ready.');
      }
    } catch (_) {
      try {
        await _bridge.stop();
      } catch (_) {
        // The caller will reconcile server availability; task checks server too.
      }
      rethrow;
    }
  }

  @override
  Future<void> stop() async {
    if (await _bridge.isRunning()) await _bridge.stop();
  }
}

class FlutterPresenceTaskBridge implements PresenceTaskBridge {
  @override
  Future<bool> isRunning() async =>
      Platform.isAndroid && await FlutterForegroundTask.isRunningService;

  @override
  Future<String?> readyUserId() async =>
      FlutterForegroundTask.getData<String>(key: _readyUserKey);

  @override
  Future<void> start(String userId) async {
    if (!Platform.isAndroid) {
      throw UnsupportedError('Background Driver presence requires Android.');
    }
    var permission = await FlutterForegroundTask.checkNotificationPermission();
    if (permission != NotificationPermission.granted) {
      permission = await FlutterForegroundTask.requestNotificationPermission();
    }
    if (permission != NotificationPermission.granted) {
      throw StateError('Allow Driver presence notifications to go online.');
    }
    FlutterForegroundTask.init(
      androidNotificationOptions: AndroidNotificationOptions(
        channelId: 'driver_presence',
        channelName: 'Driver online presence',
        channelDescription: 'Location updates while you are online as a Driver',
        onlyAlertOnce: true,
      ),
      iosNotificationOptions: const IOSNotificationOptions(),
      foregroundTaskOptions: ForegroundTaskOptions(
        eventAction: ForegroundTaskEventAction.repeat(25000),
        autoRunOnBoot: false,
        autoRunOnMyPackageReplaced: false,
        allowAutoRestart: false,
        stopWithTask: false,
      ),
    );
    await FlutterForegroundTask.removeData(key: _readyUserKey);
    if (!await FlutterForegroundTask.saveData(
      key: _expectedUserKey,
      value: userId,
    )) {
      throw StateError('Cannot initialize Driver presence task.');
    }
    final result = await FlutterForegroundTask.startService(
      serviceId: 302,
      serviceTypes: [ForegroundServiceTypes.location],
      notificationTitle: 'Driver is online',
      notificationText: 'Updating your location for ride requests',
      callback: driverPresenceTaskCallback,
    );
    if (result case ServiceRequestFailure(:final error)) throw error;
    // The native service being up is insufficient: prove the Dart task can
    // authenticate and publish once before claiming durable online presence.
    for (var attempt = 0; attempt < 50; attempt++) {
      if (await readyUserId() == userId) return;
      await Future<void>.delayed(const Duration(milliseconds: 100));
    }
    throw StateError('Driver presence task did not publish its first fix.');
  }

  @override
  Future<void> stop() async {
    if (await isRunning()) {
      final result = await FlutterForegroundTask.stopService();
      if (result case ServiceRequestFailure(:final error)) throw error;
    }
    await FlutterForegroundTask.removeData(key: _readyUserKey);
    await FlutterForegroundTask.removeData(key: _expectedUserKey);
  }
}

@pragma('vm:entry-point')
void driverPresenceTaskCallback() {
  DartPluginRegistrant.ensureInitialized();
  FlutterForegroundTask.setTaskHandler(_DriverPresenceTaskHandler());
}

class _DriverPresenceTaskHandler extends TaskHandler {
  DriverPresenceLoop? _loop;

  @override
  Future<void> onStart(DateTime timestamp, TaskStarter starter) async {
    WidgetsFlutterBinding.ensureInitialized();
    final userId = await FlutterForegroundTask.getData<String>(
      key: _expectedUserKey,
    );
    if (userId == null || userId.isEmpty) {
      await FlutterForegroundTask.stopService();
      return;
    }
    final sessions = SecureSessionStore(const FlutterSecureStorage());
    final repository = ApiDriverRepository(
      Dio(
        BaseOptions(
          baseUrl: AppConfig.apiBaseUrl,
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 10),
          headers: {'Accept': 'application/json'},
        ),
      ),
      sessions,
    );
    _loop = DriverPresenceLoop(
      publisher: DriverPresencePublisher(
        repository: repository,
        location: GeolocatorDeviceLocation(),
        sessions: sessions,
      ),
      driverUserId: userId,
      onStop: () async {
        await FlutterForegroundTask.stopService();
      },
    );
    final ready = await _loop!.tick();
    if (ready && await FlutterForegroundTask.isRunningService) {
      // A first successful authenticated location write is the ready signal.
      // A transient failure stays unacknowledged, so start rolls back safely.
      await FlutterForegroundTask.saveData(key: _readyUserKey, value: userId);
    }
  }

  @override
  void onRepeatEvent(DateTime timestamp) {
    final loop = _loop;
    if (loop != null) unawaited(loop.tick());
  }

  @override
  Future<void> onDestroy(DateTime timestamp, bool isTimeout) async {
    _loop = null;
    await FlutterForegroundTask.removeData(key: _readyUserKey);
  }
}
