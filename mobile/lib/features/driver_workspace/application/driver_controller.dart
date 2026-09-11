import '../domain/operating_state.dart';
import '../domain/registered_vehicle.dart';
import 'package:flutter/foundation.dart';

import '../../rider_request/data/device_location.dart';
import '../data/driver_repository.dart';
import '../domain/driver_profile.dart';

/// Serializes explicit Driver operations; availability is always server-confirmed.
class DriverController extends ChangeNotifier {
  DriverController(this._repository, this._location) {
    load();
  }
  final DriverRepository _repository;
  final DeviceLocation _location;
  DriverProfile? profile;
  OperatingState? operation;
  List<RegisteredVehicle> vehicles = [];
  PublishedDriverLocation? location;
  bool loaded = false;
  bool busy = false;
  String? error;
  bool _disposed = false;

  Future<void> _run(Future<void> Function() action) async {
    if (busy || _disposed) return;
    busy = true;
    error = null;
    notifyListeners();
    try {
      await action();
    } catch (failure) {
      error = '$failure';
    } finally {
      busy = false;
      if (!_disposed) notifyListeners();
    }
  }

  Future<void> load() => _run(() async {
    profile = await _repository.get();
    operation = null;
    vehicles = [];
    if (profile != null) {
      vehicles = await _repository.listVehicles();
      operation = await _repository.operatingState();
    }
    loaded = true;
  });

  Future<void> selectOperation(String vehicleId, String serviceCode) => _run(() async {
    operation = await _repository.selectOperation(vehicleId, serviceCode);
  });

  Future<void> onboard(String name, DriverVehicle vehicle) => _run(() async {
    profile = await _repository.onboard(name, vehicle);
    loaded = true;
  });

  Future<void> _publish() async {
    final point = await _location.current().timeout(
      const Duration(seconds: 20),
    );
    if (_disposed) return;
    location = await _repository.publishLocation(point);
  }

  Future<void> publishLocation() => _run(() async {
    if (profile == null) return;
    await _publish();
  });

  Future<void> setOnline(bool online) => _run(() async {
    if (profile == null) return;
    // A failed location update must not turn an offline Driver online.
    if (online && operation?.valid != true) { throw StateError('Select an approved vehicle and service first.'); }
    if (online) await _publish();
    if (_disposed) return;
    profile = await _repository.setOnline(online);
    operation = await _repository.operatingState();
  });

  @override
  void dispose() {
    _disposed = true;
    super.dispose();
  }
}
