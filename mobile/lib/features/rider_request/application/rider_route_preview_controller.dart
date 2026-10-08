import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../../core/network/api_exception.dart';
import '../data/route_preview_repository.dart';
import '../domain/ride_request.dart';
import '../domain/route_preview.dart';
import 'rider_request_controller.dart';

@immutable
class RiderRoutePreviewState {
  const RiderRoutePreviewState({
    this.preview,
    this.loading = false,
    this.error,
  });

  final RoutePreview? preview;
  final bool loading;
  final String? error;
}

class RiderRoutePreviewController extends ChangeNotifier {
  RiderRoutePreviewController(this._repository, this._rider) {
    _rider.addListener(_selectionChanged);
    _selectionChanged();
  }

  final RoutePreviewRepository _repository;
  final RiderRequestController _rider;

  RiderRoutePreviewState _state = const RiderRoutePreviewState();
  RiderRoutePreviewState get state => _state;

  String? _selectionKey;
  String? get selectionKey => _selectionKey;
  int _revision = 0;
  bool _disposed = false;

  void _selectionChanged() {
    if (_disposed) {
      return;
    }

    final pickup = _rider.state.pickup;
    final pickupPlaceId = _rider.state.pickupPlaceId;
    final destination = _rider.state.destination;
    final destinationPlaceId = _rider.state.destinationPlaceId;
    final serviceCode = _rider.serviceCode.trim();
    final key = _key(
      pickup,
      pickupPlaceId,
      destination,
      destinationPlaceId,
      serviceCode,
    );
    if (key == _selectionKey) {
      return;
    }

    _selectionKey = key;
    _repository.cancel();
    final revision = ++_revision;

    if (pickup == null || destination == null || serviceCode.isEmpty) {
      _set(const RiderRoutePreviewState());
      return;
    }

    _set(const RiderRoutePreviewState(loading: true));
    unawaited(
      _load(
        pickup: pickup,
        pickupPlaceId: pickupPlaceId,
        destination: destination,
        destinationPlaceId: destinationPlaceId,
        serviceCode: serviceCode,
        revision: revision,
      ),
    );
  }

  Future<void> retry() async {
    final pickup = _rider.state.pickup;
    final pickupPlaceId = _rider.state.pickupPlaceId;
    final destination = _rider.state.destination;
    final destinationPlaceId = _rider.state.destinationPlaceId;
    final serviceCode = _rider.serviceCode.trim();
    if (pickup == null || destination == null || serviceCode.isEmpty) {
      return;
    }

    _repository.cancel();
    final revision = ++_revision;
    _set(const RiderRoutePreviewState(loading: true));
    await _load(
      pickup: pickup,
      pickupPlaceId: pickupPlaceId,
      destination: destination,
      destinationPlaceId: destinationPlaceId,
      serviceCode: serviceCode,
      revision: revision,
    );
  }

  Future<void> _load({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
    required int revision,
  }) async {
    try {
      final preview = await _repository.preview(
        pickup: pickup,
        pickupPlaceId: pickupPlaceId,
        destination: destination,
        destinationPlaceId: destinationPlaceId,
        serviceCode: serviceCode,
      );
      if (!_isCurrent(revision)) {
        return;
      }
      _set(RiderRoutePreviewState(preview: preview));
    } on RoutePreviewCancelled {
      // A newer pickup, destination, service, or explicit retry superseded it.
    } catch (error) {
      if (!_isCurrent(revision)) {
        return;
      }
      _set(RiderRoutePreviewState(error: _message(error)));
    }
  }

  bool _isCurrent(int revision) => !_disposed && revision == _revision;

  String _key(
    GeoPoint? pickup,
    String? pickupPlaceId,
    GeoPoint? destination,
    String? destinationPlaceId,
    String serviceCode,
  ) =>
      '${pickup?.latitude},${pickup?.longitude},${pickupPlaceId ?? ''}|'
      '${destination?.latitude},${destination?.longitude},'
      '${destinationPlaceId ?? ''}|$serviceCode';

  String _message(Object error) {
    if (error is ApiException) {
      return switch (error.statusCode) {
        404 => 'No driving route was found between these locations.',
        503 => 'Route preview is temporarily unavailable.',
        _ => error.message,
      };
    }
    if (error is FormatException) {
      return 'The route provider returned an invalid route.';
    }
    return 'Unable to load the driving route.';
  }

  void _set(RiderRoutePreviewState value) {
    if (_disposed) {
      return;
    }
    _state = value;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    _rider.removeListener(_selectionChanged);
    _repository.cancel();
    super.dispose();
  }
}
