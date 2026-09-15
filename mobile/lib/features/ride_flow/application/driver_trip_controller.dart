import 'dart:async';

import 'package:flutter/foundation.dart';

import '../domain/trip.dart';
import '../ride_flow_repository.dart';

class DriverTripController extends ChangeNotifier {
  DriverTripController(
    this.repository, {
    this.interval = const Duration(seconds: 5),
  }) {
    refresh();
  }

  final RideFlowRepository repository;
  final Duration interval;

  TripSnapshot? trip;
  List<DriverTripHistoryItem> history = [];
  bool busy = false;
  bool loaded = false;
  String? error;

  bool _disposed = false;
  bool _foreground = true;
  Timer? _timer;
  Completer<void>? _idle;

  String? get status => trip?.status;

  Future<void> _load() async {
    final current = await repository.getCurrentDriverTrip();
    final trips = await repository.listDriverTrips();

    if (_disposed) return;
    trip = current;
    history = trips;
    loaded = true;
  }

  Future<void> refresh() => _run(_load);

  Future<void> startTrip(String rideRequestId) =>
      _command(() => repository.startTrip(rideRequestId));

  Future<void> completeTrip(String rideRequestId) =>
      _command(() => repository.completeTrip(rideRequestId));

  Future<void> confirmCashCollected(String rideRequestId) =>
      _command(() => repository.confirmCashCollected(rideRequestId));

  Future<void> cancelTrip(String rideRequestId) =>
      _command(() => repository.cancelDriverTrip(rideRequestId));

  Future<void> _command(Future<void> Function() command) => _run(() async {
    try {
      await command();
    } catch (_) {
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
    if (!_disposed && _foreground) {
      _timer = Timer(interval, refresh);
    }
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
