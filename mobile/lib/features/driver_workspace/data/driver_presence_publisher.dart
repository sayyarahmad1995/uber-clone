import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';
import '../../rider_request/data/device_location.dart';
import 'driver_repository.dart';

/// Renews the existing server lease only for the original authenticated Driver.
class DriverPresencePublisher {
  DriverPresencePublisher({
    required DriverRepository repository,
    required DeviceLocation location,
    required SessionStore sessions,
  }) : this._(repository, location, sessions);

  DriverPresencePublisher._(this._repository, this._location, this._sessions);

  final DriverRepository _repository;
  final DeviceLocation _location;
  final SessionStore _sessions;

  /// False means this task must stop; transient failures remain retryable.
  Future<bool> renew(String expectedDriverUserId) async {
    if (await _sessions.readValidToken() == null) return false;
    try {
      final profile = await _repository.get();
      if (profile == null ||
          profile.userId != expectedDriverUserId ||
          !profile.isOnline) {
        return false;
      }
      final point = await _location.current();
      await _repository.publishLocation(point);
      return true;
    } on ApiException catch (error) {
      if (error.isUnauthorized) return false;
      rethrow;
    }
  }
}

/// Serializes service ticks; a missed tick never fabricates a fresh lease.
class DriverPresenceLoop {
  DriverPresenceLoop({
    required DriverPresencePublisher publisher,
    required String driverUserId,
    required Future<void> Function() onStop,
  }) : this._(publisher, driverUserId, onStop);

  DriverPresenceLoop._(this._publisher, this._driverUserId, this._onStop);

  final DriverPresencePublisher _publisher;
  final String _driverUserId;
  final Future<void> Function() _onStop;
  bool _inFlight = false;
  bool _stopped = false;

  Future<bool> tick() async {
    if (_inFlight) return false;
    if (_stopped) {
      try {
        await _onStop();
      } catch (_) {
        // Retry on the next native event; never resume presence publication.
      }
      return false;
    }
    _inFlight = true;
    try {
      if (!await _publisher.renew(_driverUserId)) {
        _stopped = true;
        await _onStop();
        return false;
      }
      return true;
    } catch (_) {
      // A failed fix or request must not report fabricated freshness.
      // The next tick may retry; server expiry remains authoritative.
      return false;
    } finally {
      _inFlight = false;
    }
  }
}
