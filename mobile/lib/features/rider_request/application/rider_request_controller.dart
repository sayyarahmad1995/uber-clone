import 'package:flutter/foundation.dart';

import '../data/device_location.dart';
import '../data/ride_request_repository.dart';
import '../domain/ride_request.dart';

@immutable
class RiderRequestState {
  const RiderRequestState({
    this.requests = const [],
    this.pickup,
    this.destination,
    this.loading = true,
    this.submitting = false,
    this.locating = false,
    this.error,
  });
  final List<RideRequest> requests;
  final GeoPoint? pickup;
  final GeoPoint? destination;
  final bool loading;
  final bool submitting;
  final bool locating;
  final String? error;

  RideRequest? get active {
    for (final request in requests) {
      if (request.status != 'cancelled' &&
          request.trip?.status != 'completed' &&
          request.trip?.status != 'cancelled') {
        return request;
      }
    }
    return null;
  }

  RiderRequestState copyWith({
    List<RideRequest>? requests,
    GeoPoint? pickup,
    GeoPoint? destination,
    bool? loading,
    bool? submitting,
    bool? locating,
    String? error,
    bool clearError = false,
  }) => RiderRequestState(
    requests: requests ?? this.requests,
    pickup: pickup ?? this.pickup,
    destination: destination ?? this.destination,
    loading: loading ?? this.loading,
    submitting: submitting ?? this.submitting,
    locating: locating ?? this.locating,
    error: clearError ? null : error ?? this.error,
  );
}

class RiderRequestController extends ChangeNotifier {
  RiderRequestController(this._repository, this._location) {
    load();
  }
  final RideRequestRepository _repository;
  final DeviceLocation _location;
  RiderRequestState _state = const RiderRequestState();
  RiderRequestState get state => _state;

  Future<void> load() async {
    _set(_state.copyWith(loading: true, clearError: true));
    try {
      final requests = await _repository.list();
      _set(_state.copyWith(requests: requests, loading: false));
    } catch (error) {
      _set(_state.copyWith(loading: false, error: '$error'));
    }
  }

  Future<void> useCurrentPickup() async {
    _set(_state.copyWith(locating: true, clearError: true));
    try {
      final pickup = await _location.current();
      _set(_state.copyWith(pickup: pickup, locating: false));
    } catch (error) {
      _set(_state.copyWith(locating: false, error: '$error'));
    }
  }

  void setPickup(GeoPoint point) =>
      _set(_state.copyWith(pickup: point, clearError: true));
  void setDestination(GeoPoint point) =>
      _set(_state.copyWith(destination: point, clearError: true));

  Future<bool> submit({
    required int amountMinor,
    required String currency,
  }) async {
    final pickup = _state.pickup;
    final destination = _state.destination;
    if (pickup == null || destination == null) {
      _set(_state.copyWith(error: 'Choose both pickup and destination.'));
      return false;
    }
    _set(_state.copyWith(submitting: true, clearError: true));
    try {
      final request = await _repository.create(
        pickup: pickup,
        destination: destination,
        proposedFare: Money(amountMinor: amountMinor, currency: currency),
      );
      _set(
        _state.copyWith(
          requests: [request, ..._state.requests],
          submitting: false,
        ),
      );
      return true;
    } catch (error) {
      _set(_state.copyWith(submitting: false, error: '$error'));
      return false;
    }
  }

  Future<void> refreshActive() async {
    final active = _state.active;
    if (active == null) return load();
    _set(_state.copyWith(loading: true, clearError: true));
    try {
      final updated = await _repository.get(active.id);
      final requests = [
        updated,
        ..._state.requests.where((item) => item.id != updated.id),
      ];
      _set(_state.copyWith(requests: requests, loading: false));
    } catch (error) {
      _set(_state.copyWith(loading: false, error: '$error'));
    }
  }

  Future<void> cancelActive() async {
    final active = _state.active;
    if (active == null) return;
    _set(_state.copyWith(submitting: true, clearError: true));
    try {
      await _repository.cancel(active.id);
      final requests = await _repository.list();
      _set(_state.copyWith(requests: requests, submitting: false));
    } catch (error) {
      _set(_state.copyWith(submitting: false, error: '$error'));
    }
  }

  void _set(RiderRequestState value) {
    _state = value;
    notifyListeners();
  }
}
