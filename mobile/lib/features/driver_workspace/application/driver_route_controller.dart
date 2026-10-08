import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../../core/network/api_exception.dart';
import '../../ride_flow/application/driver_trip_controller.dart';
import '../../ride_flow/domain/trip.dart';
import '../../rider_request/data/device_location.dart';
import '../../rider_request/domain/ride_request.dart' show GeoPoint;
import '../../rider_request/domain/route_preview.dart';
import '../data/driver_route_repository.dart';

@immutable
class DriverRouteState {
  const DriverRouteState({
    this.preview,
    this.status,
    this.loading = false,
    this.error,
  });

  final RoutePreview? preview;
  final String? status;
  final bool loading;
  final String? error;
}

class DriverRouteController extends ChangeNotifier {
  DriverRouteController(this._repository, this._flow, this._location) {
    _flow.addListener(_tripChanged);
    _tripChanged();
  }

  final DriverRouteRepository _repository;
  final DriverTripController _flow;
  final DeviceLocation _location;
  DriverRouteState _state = const DriverRouteState();
  DriverRouteState get state => _state;
  String? _selectionKey;
  int _revision = 0;
  bool _disposed = false;

  bool _routable(TripSnapshot? trip) =>
      trip != null &&
      trip.rideRequestId != null &&
      (trip.status == 'assigned' || trip.status == 'in_progress');

  String? _key(TripSnapshot? trip) {
    if (!_routable(trip)) return null;
    return '${trip!.rideRequestId}|${trip.status}|'
        '${trip.pickup?.latitude},${trip.pickup?.longitude}|'
        '${trip.destination?.latitude},${trip.destination?.longitude}';
  }

  void _tripChanged() {
    if (_disposed) return;
    final trip = _flow.trip;
    final key = _key(trip);
    if (key == _selectionKey) return;
    _selectionKey = key;
    unawaited(_loadTrip(trip));
  }

  Future<void> retry() => _loadTrip(_flow.trip);

  Future<void> _loadTrip(TripSnapshot? trip) async {
    if (_disposed) return;
    _repository.cancel();
    final revision = ++_revision;
    if (!_routable(trip)) {
      _set(const DriverRouteState());
      return;
    }
    final selected = trip!;
    _set(DriverRouteState(status: selected.status, loading: true));
    try {
      // Assignment needs a live fix; an in-progress route uses server endpoints.
      final GeoPoint? origin = selected.status == 'assigned'
          ? await _location.current().timeout(const Duration(seconds: 20))
          : null;
      if (!_current(revision)) return;
      final preview = await _repository.preview(
        rideRequestId: selected.rideRequestId!,
        status: selected.status,
        origin: origin,
      );
      if (!_current(revision)) return;
      _set(DriverRouteState(status: selected.status, preview: preview));
    } on DriverRouteCancelled {
      // A stage change, account change, retry or disposal superseded this reply.
    } catch (error) {
      if (!_current(revision)) return;
      _set(DriverRouteState(status: selected.status, error: _message(error)));
    }
  }

  bool _current(int revision) => !_disposed && revision == _revision;

  String _message(Object error) {
    if (error is LocationUnavailable) return error.message;
    if (error is TimeoutException) {
      return 'Unable to get your current location.';
    }
    if (error is ApiException) {
      return switch (error.statusCode) {
        404 => 'No driving route or active trip was found.',
        409 => 'The trip changed. Refresh the trip and retry the route.',
        503 => 'Driving route is temporarily unavailable.',
        _ => 'Unable to load the driving route.',
      };
    }
    return 'Unable to load the driving route.';
  }

  void _set(DriverRouteState state) {
    if (_disposed) return;
    _state = state;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    _flow.removeListener(_tripChanged);
    _repository.cancel();
    super.dispose();
  }
}
