import 'package:flutter/foundation.dart';

import 'polling_loop.dart';
import '../domain/marketplace_request.dart';
import '../ride_flow_repository.dart';

class DriverMarketplaceController extends ChangeNotifier {
  DriverMarketplaceController(
    this.repository, {
    this.interval = const Duration(seconds: 5),
  }) {
    refresh();
  }

  final RideFlowRepository repository;
  final Duration interval;
  late final PollingLoop _polling = PollingLoop(interval: interval);

  List<MarketplaceRequest> requests = [];
  bool loaded = false;
  String? error;

  bool _disposed = false;

  bool get busy => _polling.busy;

  Future<void> _load() async {
    final value = await repository.listMarketplaceRequests();
    if (_disposed) return;

    requests = value;
    loaded = true;
  }

  Future<void> refresh() {
    if (_polling.busy) return Future<void>.value();
    return _polling.runRefresh(_refreshAndReport);
  }

  Future<void> submitOffer(String rideRequestId, int amountMinor) =>
      _command(() => repository.submitOffer(rideRequestId, amountMinor));

  Future<void> acceptProposedFare(String rideRequestId) =>
      _command(() => repository.acceptProposedFare(rideRequestId));

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
