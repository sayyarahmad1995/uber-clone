import 'dart:async';

import 'package:flutter/foundation.dart';

import '../domain/ride_offer.dart';
import '../domain/ride_snapshot.dart';
import '../ride_flow_repository.dart';

class RiderActiveRideController extends ChangeNotifier {
  RiderActiveRideController(
    this.repository, {
    required this.rideId,
    this.interval = const Duration(seconds: 5),
  }) {
    refresh();
  }

  final RideFlowRepository repository;
  final String rideId;
  final Duration interval;

  RiderRideSnapshot? riderRide;
  DriverLocationSnapshot? location;
  List<RiderOfferComparison> offers = [];
  bool busy = false;
  bool loaded = false;
  String? error;

  bool _disposed = false;
  bool _foreground = true;
  Timer? _timer;
  Completer<void>? _idle;

  String? get status => riderRide?.trip?.status ?? riderRide?.status;

  Future<void> _load() async {
    final ride = await repository.getRiderRide(rideId);
    final active =
        ride.trip != null &&
        !['completed', 'cancelled'].contains(ride.trip!.status);

    final comparison = ride.trip == null && ride.status == 'requested'
        ? await repository.listRiderOffers(rideId)
        : <RiderOfferComparison>[];

    DriverLocationSnapshot? position;
    if (active) {
      try {
        position = await repository.getDriverLocation(rideId);
      } catch (_) {
        error = 'Driver location is unavailable. Trip status is up to date.';
      }
    }

    if (_disposed) return;
    riderRide = ride;
    offers = comparison;
    location = position;
    loaded = true;
  }

  Future<void> refresh() => _run(_load);

  Future<void> selectOffer(String driverUserId, DateTime updatedAt) =>
      _command(() => repository.selectOffer(rideId, driverUserId, updatedAt));

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
