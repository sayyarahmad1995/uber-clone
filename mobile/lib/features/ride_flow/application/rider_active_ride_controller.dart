import 'package:flutter/foundation.dart';

import 'polling_loop.dart';
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
  late final PollingLoop _polling = PollingLoop(interval: interval);

  RiderRideSnapshot? riderRide;
  DriverLocationSnapshot? location;
  List<RiderOfferComparison> offers = [];
  bool loaded = false;
  String? error;

  bool _disposed = false;

  bool get busy => _polling.busy;

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

  Future<void> refresh() {
    if (_polling.busy) return Future<void>.value();
    return _polling.runRefresh(_refreshAndReport);
  }

  Future<void> selectOffer(String driverUserId, DateTime updatedAt) =>
      _command(() => repository.selectOffer(rideId, driverUserId, updatedAt));

  Future<void> declineOffer(String driverUserId) =>
      _command(() => repository.declineOffer(rideId, driverUserId));

  Future<void> _command(Future<void> Function() command) async {
    var started = false;
    try {
      await _polling.runCommand(() async {
        started = true;
        error = null;
        notifyListeners();
        await command();
      }, reload: _load);
    } catch (e) {
      error = '$e';
    } finally {
      if (started && !_disposed) notifyListeners();
    }
  }

  void setForeground(bool value) {
    _polling.setForeground(value, _refreshAndReport);
  }

  Future<void> _refreshAndReport() async {
    if (_disposed) return;
    error = null;
    notifyListeners();
    try {
      await _load();
    } catch (e) {
      error = '$e';
    } finally {
      if (!_disposed) notifyListeners();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _polling.dispose();
    super.dispose();
  }
}
