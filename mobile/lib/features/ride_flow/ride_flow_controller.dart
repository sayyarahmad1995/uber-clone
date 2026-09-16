import 'dart:async';

import 'package:flutter/foundation.dart';

import 'ride_flow_repository.dart';

/// Server-owned state. Polling is serialized with commands and stopped in background.
class RideFlowController extends ChangeNotifier {
  RideFlowController(
    this.repository, {
    required this.mode,
    this.rideId,
    this.interval = const Duration(seconds: 5),
  }) {
    refresh();
  }
  final RideFlowRepository repository;
  final RideFlowMode mode;
  final String? rideId;
  final Duration interval;
  bool get driver => mode == RideFlowMode.driver;
  RideFlowPayload? current;
  RideFlowPayload? location;
  List<RideFlowPayload> offers = [];
  List<RideFlowPayload> requests = [];
  List<RideFlowPayload> history = [];
  bool busy = false;
  bool loaded = false;
  String? error;
  bool _disposed = false;
  bool _foreground = true;
  Timer? _timer;
  Completer<void>? _idle;
  String? get status {
    if (driver) return current?['status'] as String?;
    return (current?['trip']?['status'] ?? current?['status']) as String?;
  }

  List<RideFlowPayload> _list(RideFlowPayload json, String key) =>
      (json[key] as List? ?? [])
          .map((e) => Map<String, dynamic>.from(e as Map))
          .toList();

  Future<void> _load() async {
    if (driver) {
      final trip = await repository.loadDriverTrip();
      if (_disposed) return;
      current = trip;
      loaded = true;
      final feed = trip == null
          ? await repository.loadDriverOpportunities()
          : <String, dynamic>{};
      final past = await repository.loadDriverHistory();
      if (_disposed) return;
      current = trip;
      requests = _list(feed, 'ride_requests');
      history = _list(past, 'trips');
    } else {
      final ride = await repository.loadRiderRide(rideId!);
      final active =
          ride['trip'] != null &&
          !['completed', 'cancelled'].contains(ride['trip']['status']);
      final comparison = ride['trip'] == null && ride['status'] == 'requested'
          ? await repository.loadRiderOffers(rideId!)
          : <String, dynamic>{};
      RideFlowPayload? position;
      if (active) {
        try {
          position = await repository.loadDriverLocation(rideId!);
        } catch (_) {
          error = 'Driver location is unavailable. Trip status is up to date.';
        }
      }
      if (_disposed) return;
      current = ride;
      offers = _list(comparison, 'offers');
      location = position;
    }
    loaded = true;
  }

  Future<void> refresh() => _run(_load);

  Future<void> runCommand(Future<void> Function() command) => _run(() async {
    try {
      await command();
    } catch (_) {
      // A response may be lost after commit, or the offer may have changed.
      // Reload authoritative state without claiming that a failed command succeeded.
      try {
        await _load();
      } catch (_) {}
      rethrow;
    }
    await _load();
  }, waitForBusy: true);

  Future<void> _run(
    Future<void> Function() task, {
    bool waitForBusy = false,
  }) async {
    if (_disposed || !_foreground) return;
    if (waitForBusy) {
      while (busy && !_disposed && _foreground) {
        final idle = _idle;
        if (idle == null) break;
        await idle.future;
      }
    } else if (busy) {
      return;
    }
    if (_disposed || !_foreground) return;
    _timer?.cancel();
    final idle = Completer<void>();
    _idle = idle;
    busy = true;
    error = null;
    notifyListeners();
    try {
      await task();
    } catch (e) {
      error = '$e';
    } finally {
      busy = false;
      if (!idle.isCompleted) idle.complete();
      if (identical(_idle, idle)) _idle = null;
      if (!_disposed) {
        notifyListeners();
        _schedule();
      }
    }
  }

  void _schedule() {
    _timer?.cancel();
    if (!_disposed && _foreground) _timer = Timer(interval, refresh);
  }

  void setForeground(bool value) {
    _foreground = value;
    _timer?.cancel();
    if (value) refresh();
  }

  @override
  void dispose() {
    _disposed = true;
    _timer?.cancel();
    super.dispose();
  }
}

enum RideFlowMode { rider, driver }
