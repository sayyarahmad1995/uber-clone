import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../rider_request/data/device_location.dart';
import 'driver_presence_service.dart';
import '../data/driver_repository.dart';
import '../domain/driver_profile.dart';
import '../domain/operating_state.dart';
import '../domain/registered_vehicle.dart';

/// Serializes explicit Driver operations; availability is always server-confirmed.
class DriverController extends ChangeNotifier {
  DriverController(this._repository, this._location, this._presence) {
    load();
  }

  final DriverRepository _repository;
  final DeviceLocation _location;
  final DriverPresenceService _presence;
  DriverProfile? profile;
  OperatingState? operation;
  List<RegisteredVehicle> vehicles = [];
  PublishedDriverLocation? location;
  bool hasActiveTrip = false;
  bool onlinePresenceReady = false;

  void setActiveTrip(bool active) {
    if (hasActiveTrip == active) return;
    hasActiveTrip = active;
    _scheduleHeartbeat();
    unawaited(load());
  }

  bool loaded = false;
  bool busy = false;
  String? error;
  bool _disposed = false;
  bool _foreground = true;
  bool _entryPublishInFlight = false;
  Timer? _heartbeat;

  void setForeground(bool foreground) {
    _foreground = foreground;
    if (!foreground) {
      _heartbeat?.cancel();
    } else {
      unawaited(load());
    }
  }

  void _scheduleHeartbeat() {
    _heartbeat?.cancel();
    if (_disposed ||
        !_foreground ||
        !hasActiveTrip ||
        profile?.isOnline == true) {
      return;
    }
    _heartbeat = Timer(const Duration(seconds: 20), () {
      if (busy) {
        _scheduleHeartbeat();
        return;
      }
      unawaited(
        _run(() async {
          await _publish();
          if (_disposed || !_foreground) return;
          profile = await _repository.get();
          operation = await _repository.operatingState();
        }),
      );
    });
  }

  Future<void> _run(
    Future<void> Function() action, {
    bool reportError = true,
  }) async {
    if (busy || _disposed) return;
    busy = true;
    error = null;
    notifyListeners();
    try {
      await action();
    } catch (failure) {
      if (reportError) {
        error = '$failure';
      }
    } finally {
      busy = false;
      _scheduleHeartbeat();
      if (!_disposed) notifyListeners();
    }
  }

  Future<void> load() async {
    await _run(() async {
      profile = await _repository.get();
      operation = null;
      vehicles = [];
      if (profile != null) {
        vehicles = await _repository.listVehicles();
        operation = await _repository.operatingState();
        loaded = true;
        if (_foreground && profile!.isOnline) {
          await _ensurePresence(profile!.userId);
        } else if (!profile!.isOnline &&
            await _presence.isRunningFor(profile!.userId)) {
          await _presence.stop();
        }
        if (!profile!.isOnline) onlinePresenceReady = false;
      }
      loaded = true;
    });
    if (!_disposed &&
        _foreground &&
        profile != null &&
        operation?.valid == true &&
        location == null) {
      // Background presence writes independently of this UI controller. After
      // a process restart an online Driver can therefore have durable presence
      // but no in-memory location for the map. Publish one foreground fix to
      // hydrate the dashboard and trigger the shared map's initial auto-center.
      unawaited(_publishForDriverModeEntry());
    }
  }

  Future<void> selectOperation(String vehicleId) => _run(() async {
    operation = await _repository.selectOperation(vehicleId);
  });

  Future<void> _publish() async {
    final point = await _location.current().timeout(
      const Duration(seconds: 20),
    );
    if (_disposed || !_foreground) return;
    location = await _repository.publishLocation(point);
  }

  Future<void> _publishForDriverModeEntry() async {
    if (_entryPublishInFlight ||
        _disposed ||
        !_foreground ||
        profile == null ||
        operation?.valid != true) {
      return;
    }
    _entryPublishInFlight = true;
    try {
      await _publish();
    } catch (_) {
      // Driver mode entry should not surface permission or location errors.
    } finally {
      _entryPublishInFlight = false;
      if (!_disposed) notifyListeners();
    }
  }

  Future<void> publishLocation() => _run(() async {
    if (profile == null) return;
    await _publish();
  });

  Future<void> setOnline(bool online) => _run(() async {
    if (profile == null) return;
    // A failed location update must not turn an offline Driver online.
    if (online && operation?.valid != true) {
      throw StateError(
        'Select an approved vehicle with an available service first.',
      );
    }
    if (online && profile!.isOnline) {
      await _ensurePresence(profile!.userId);
      return;
    }
    if (online) await _publish();
    if (_disposed || (online && !_foreground)) return;
    profile = await _repository.setOnline(online);
    if (online) {
      await _ensurePresence(profile!.userId);
    } else {
      onlinePresenceReady = false;
      await _presence.stop();
    }
    operation = await _repository.operatingState();
  });

  Future<void> _ensurePresence(String driverUserId) async {
    if (await _presence.isRunningFor(driverUserId)) {
      onlinePresenceReady = true;
      return;
    }
    try {
      await _presence.start(driverUserId);
      onlinePresenceReady = true;
    } catch (failure) {
      onlinePresenceReady = false;
      Object? rollbackFailure;
      try {
        profile = await _repository.setOnline(false);
      } catch (error) {
        rollbackFailure = error;
      } finally {
        try {
          await _presence.stop();
        } catch (_) {
          // Task checks server state before any location request.
        }
      }
      throw StateError(
        'Driver presence unavailable: $failure'
        '${rollbackFailure == null ? '' : '; server rollback failed: $rollbackFailure'}',
      );
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _heartbeat?.cancel();
    super.dispose();
  }
}
