import 'package:flutter/foundation.dart';

import '../data/driver_onboarding_repository.dart';
import '../domain/driver_onboarding.dart';
import '../domain/driver_profile.dart';

class DriverOnboardingController extends ChangeNotifier {
  DriverOnboardingController(this._repository) {
    load();
  }

  final DriverOnboardingRepository _repository;
  List<DriverServiceOption> services = const [];
  DriverOnboardingApplication? application;
  bool loaded = false;
  bool busy = false;
  String? error;
  bool _disposed = false;

  Future<void> load() async {
    if (busy || _disposed) return;
    busy = true;
    error = null;
    notifyListeners();
    try {
      final results = await Future.wait<dynamic>([
        _repository.listServices(),
        _repository.getLatest(),
      ]);
      services = results[0] as List<DriverServiceOption>;
      application = results[1] as DriverOnboardingApplication?;
      loaded = true;
    } catch (failure) {
      error = '$failure';
      loaded = true;
    } finally {
      busy = false;
      if (!_disposed) notifyListeners();
    }
  }

  Future<DriverOnboardingPrecheck?> precheck({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async {
    if (busy || _disposed) return null;
    busy = true;
    error = null;
    notifyListeners();
    try {
      return await _repository.precheck(
        displayName: displayName,
        serviceCode: serviceCode,
        vehicle: vehicle,
      );
    } catch (failure) {
      error = '$failure';
      return null;
    } finally {
      busy = false;
      if (!_disposed) notifyListeners();
    }
  }

  Future<void> submit({
    required String displayName,
    required String serviceCode,
    required DriverVehicle vehicle,
  }) async {
    if (busy || _disposed) return;
    busy = true;
    error = null;
    notifyListeners();
    try {
      application = await _repository.submit(
        displayName: displayName,
        serviceCode: serviceCode,
        vehicle: vehicle,
      );
    } catch (failure) {
      error = '$failure';
    } finally {
      busy = false;
      if (!_disposed) notifyListeners();
    }
  }

  void startNewApplication() {
    if (busy || _disposed) return;
    application = null;
    error = null;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    super.dispose();
  }
}
