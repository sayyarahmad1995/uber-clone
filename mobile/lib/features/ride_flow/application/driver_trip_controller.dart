import 'package:flutter/foundation.dart';

import 'polling_loop.dart';
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
  late final PollingLoop _polling = PollingLoop(interval: interval);

  TripSnapshot? trip;
  List<DriverTripHistoryItem> history = [];
  bool loaded = false;
  String? error;

  bool _disposed = false;

  bool get busy => _polling.busy;

  String? get status => trip?.status;

  Future<void> _load() async {
    final current = await repository.getCurrentDriverTrip();
    final trips = await repository.listDriverTrips();

    if (_disposed) return;
    trip = current;
    history = trips;
    loaded = true;
  }

  Future<void> refresh() {
    if (_polling.busy) return Future<void>.value();
    return _polling.runRefresh(_refreshAndReport);
  }

  Future<void> startTrip(String rideRequestId) =>
      _command(() => repository.startTrip(rideRequestId));

  Future<void> completeTrip(String rideRequestId) =>
      _command(() => repository.completeTrip(rideRequestId));

  Future<void> confirmCashCollected(String rideRequestId) =>
      _command(() => repository.confirmCashCollected(rideRequestId));

  Future<void> cancelTrip(String rideRequestId) =>
      _command(() => repository.cancelDriverTrip(rideRequestId));

  Future<void> _command(Future<void> Function() command) async {
    error = null;
    notifyListeners();
    try {
      await _polling.runCommand(command, reload: _load);
    } catch (e) {
      error = '$e';
    } finally {
      if (!_disposed) notifyListeners();
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
