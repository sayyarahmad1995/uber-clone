import 'dart:async';

import 'package:flutter/foundation.dart';

import 'domain/marketplace_request.dart';
import 'domain/ride_offer.dart';
import 'domain/ride_snapshot.dart';
import 'domain/trip.dart';
import 'ride_flow_repository.dart';

/// Server-owned state. Polling is serialized with commands and stopped in background.
class RideFlowController extends ChangeNotifier {
  RideFlowController(
    this.repository, {
    this.rideId,
    this.interval = const Duration(seconds: 5),
  }) {
    refresh();
  }

  final RideFlowRepository repository;
  final String? rideId;
  final Duration interval;

  bool get driver => rideId == null;

  TripSnapshot? driverTrip;
  RiderRideSnapshot? riderRide;
  DriverLocationSnapshot? location;
  List<RiderOfferComparison> offers = [];
  List<MarketplaceRequest> requests = [];
  List<DriverTripHistoryItem> history = [];
  bool busy = false;
  bool loaded = false;
  String? error;
  bool _disposed = false;
  bool _foreground = true;
  Timer? _timer;
  Completer<void>? _idle;

  String? get status => driver
      ? driverTrip?.status
      : (riderRide?.trip?.status ?? riderRide?.status);

  Future<void> _load() async {
    if (driver) {
      final trip = await repository.getCurrentDriverTrip();
      if (_disposed) return;
      driverTrip = trip;
      loaded = true;
      requests = trip == null
          ? await repository.listMarketplaceRequests()
          : const [];
      history = await repository.listDriverTrips();
      if (_disposed) return;
      driverTrip = trip;
    } else {
      final id = rideId!;
      final ride = await repository.getRiderRide(id);
      final active =
          ride.trip != null && !['completed', 'cancelled'].contains(ride.trip!.status);
      final comparison = ride.trip == null && ride.status == 'requested'
          ? await repository.listRiderOffers(id)
          : <RiderOfferComparison>[];
      DriverLocationSnapshot? position;
      if (active) {
        try {
          position = await repository.getDriverLocation(id);
        } catch (_) {
          error = 'Driver location is unavailable. Trip status is up to date.';
        }
      }
      if (_disposed) return;
      riderRide = ride;
      offers = comparison;
      location = position;
    }
    loaded = true;
  }

  Future<void> refresh() => _run(_load);

  Future<void> submitOffer(String rideRequestId, int amountMinor) => _command(
    () => repository.submitOffer(rideRequestId, amountMinor),
  );

  Future<void> acceptProposedFare(String rideRequestId) => _command(
    () => repository.acceptProposedFare(rideRequestId),
  );

  Future<void> selectOffer(
    String rideRequestId,
    String driverUserId,
    DateTime updatedAt,
  ) => _command(
    () => repository.selectOffer(rideRequestId, driverUserId, updatedAt),
  );

  Future<void> startTrip(String rideRequestId) =>
      _command(() => repository.startTrip(rideRequestId));

  Future<void> completeTrip(String rideRequestId) =>
      _command(() => repository.completeTrip(rideRequestId));

  Future<void> confirmCashCollected(String rideRequestId) =>
      _command(() => repository.confirmCashCollected(rideRequestId));

  Future<void> cancelDriverTrip(String rideRequestId) =>
      _command(() => repository.cancelDriverTrip(rideRequestId));

  Future<void> _command(Future<void> Function() command) => _run(() async {
    try {
      await command();
    } catch (_) {
      // A response may be lost after commit, or authoritative state may have changed.
      // Reload without claiming that the failed command succeeded.
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
