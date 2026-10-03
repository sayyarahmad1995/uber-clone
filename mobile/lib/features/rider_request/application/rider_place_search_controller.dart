import 'dart:async';
import 'dart:math';

import 'package:flutter/foundation.dart';

import '../data/device_location.dart';
import '../data/place_search_repository.dart';
import '../domain/place_search.dart';
import '../domain/ride_request.dart';
import 'rider_request_controller.dart';

@immutable
class PlaceFieldSearchState {
  const PlaceFieldSearchState({
    this.suggestions = const [],
    this.label,
    this.searching = false,
    this.resolving = false,
    this.searched = false,
    this.error,
  });

  final List<PlaceSuggestion> suggestions;
  final String? label;
  final bool searching;
  final bool resolving;
  final bool searched;
  final String? error;

  PlaceFieldSearchState copyWith({
    List<PlaceSuggestion>? suggestions,
    String? label,
    bool? searching,
    bool? resolving,
    bool? searched,
    String? error,
    bool clearLabel = false,
    bool clearError = false,
  }) => PlaceFieldSearchState(
    suggestions: suggestions ?? this.suggestions,
    label: clearLabel ? null : label ?? this.label,
    searching: searching ?? this.searching,
    resolving: resolving ?? this.resolving,
    searched: searched ?? this.searched,
    error: clearError ? null : error ?? this.error,
  );
}

@immutable
class RiderPlaceSearchState {
  const RiderPlaceSearchState({
    this.pickup = const PlaceFieldSearchState(),
    this.destination = const PlaceFieldSearchState(),
  });

  final PlaceFieldSearchState pickup;
  final PlaceFieldSearchState destination;

  PlaceFieldSearchState field(RiderPlaceField field) =>
      field == RiderPlaceField.pickup ? pickup : destination;

  RiderPlaceSearchState replace(
    RiderPlaceField field,
    PlaceFieldSearchState value,
  ) => RiderPlaceSearchState(
    pickup: field == RiderPlaceField.pickup ? value : pickup,
    destination: field == RiderPlaceField.destination ? value : destination,
  );
}

class RiderPlaceSearchController extends ChangeNotifier {
  RiderPlaceSearchController(
    this._repository,
    this._rider,
    this._deviceLocation,
  );

  final PlaceSearchRepository _repository;
  final RiderRequestController _rider;
  final DeviceLocation _deviceLocation;

  RiderPlaceSearchState _state = const RiderPlaceSearchState();
  RiderPlaceSearchState get state => _state;

  final Map<RiderPlaceField, Timer> _debounces = {};
  final Map<RiderPlaceField, String> _sessionTokens = {};
  final Map<RiderPlaceField, int> _revisions = {
    RiderPlaceField.pickup: 0,
    RiderPlaceField.destination: 0,
  };
  Future<GeoPoint?>? _biasFuture;
  bool _disposed = false;

  void search(RiderPlaceField field, String rawInput) {
    final input = rawInput.trim();
    final revision = _nextRevision(field);
    _debounces.remove(field)?.cancel();

    final previous = _state.field(field);
    _setField(
      field,
      previous.copyWith(
        suggestions: const [],
        searching: input.length >= 3,
        resolving: false,
        searched: false,
        clearError: true,
        clearLabel: input != previous.label,
      ),
    );

    if (input.length < 3) {
      _endSession(field);
      return;
    }

    final sessionToken = _sessionTokens.putIfAbsent(field, _newSessionToken);
    _debounces[field] = Timer(
      const Duration(milliseconds: 300),
      () => unawaited(
        _autocomplete(
          field: field,
          input: input,
          sessionToken: sessionToken,
          revision: revision,
        ),
      ),
    );
  }

  Future<PlaceSelection?> select(
    RiderPlaceField field,
    PlaceSuggestion suggestion,
  ) async {
    final sessionToken = _sessionTokens[field];
    if (sessionToken == null) return null;

    _repository.cancelSession(sessionToken);
    _sessionTokens.remove(field);
    _debounces.remove(field)?.cancel();
    final revision = _nextRevision(field);
    final previous = _state.field(field);
    _setField(
      field,
      previous.copyWith(
        suggestions: const [],
        searching: false,
        resolving: true,
        searched: false,
        clearError: true,
      ),
    );

    try {
      final selected = await _repository.details(
        placeId: suggestion.placeId,
        sessionToken: sessionToken,
      );
      if (!_isCurrent(field, revision)) return null;
      _applyPoint(field, selected.point);
      _setField(field, PlaceFieldSearchState(label: selected.label));
      return selected;
    } catch (error) {
      if (!_isCurrent(field, revision)) return null;
      _setField(
        field,
        previous.copyWith(
          suggestions: const [],
          searching: false,
          resolving: false,
          searched: false,
          error: _message(error),
        ),
      );
      return null;
    }
  }

  Future<PlaceSelection?> selectPlaceId(
    RiderPlaceField field,
    String placeId,
  ) async {
    _endSession(field);
    _debounces.remove(field)?.cancel();
    final revision = _nextRevision(field);
    final previous = _state.field(field);
    _setField(
      field,
      previous.copyWith(
        suggestions: const [],
        searching: false,
        resolving: true,
        searched: false,
        clearError: true,
      ),
    );

    try {
      final selected = await _repository.placeById(placeId);
      if (!_isCurrent(field, revision)) return null;
      _applyPoint(field, selected.point);
      _setField(field, PlaceFieldSearchState(label: selected.label));
      return selected;
    } catch (error) {
      if (!_isCurrent(field, revision)) return null;
      _setField(
        field,
        previous.copyWith(
          suggestions: const [],
          searching: false,
          resolving: false,
          searched: false,
          error: _message(error),
        ),
      );
      return null;
    }
  }

  Future<PlaceSelection?> reconcilePin(
    RiderPlaceField field,
    GeoPoint point,
  ) async {
    _endSession(field);
    _debounces.remove(field)?.cancel();
    final revision = _nextRevision(field);
    _setField(field, const PlaceFieldSearchState(resolving: true));

    try {
      final resolved = await _repository.reverseGeocode(point);
      if (!_isCurrent(field, revision)) return null;
      if (resolved == null) {
        _setField(
          field,
          const PlaceFieldSearchState(
            error: 'No readable address was found for this pin.',
          ),
        );
        return null;
      }
      final selection = PlaceSelection(
        placeId: resolved.placeId,
        label: resolved.label,
        point: point,
      );
      _applyPoint(field, point);
      _setField(field, PlaceFieldSearchState(label: selection.label));
      return selection;
    } catch (error) {
      if (!_isCurrent(field, revision)) return null;
      _setField(field, PlaceFieldSearchState(error: _message(error)));
      return null;
    }
  }

  Future<void> _autocomplete({
    required RiderPlaceField field,
    required String input,
    required String sessionToken,
    required int revision,
  }) async {
    try {
      final suggestions = await _repository.autocomplete(
        input: input,
        sessionToken: sessionToken,
        bias: await _loadBias(),
      );
      if (!_isCurrent(field, revision)) return;
      final previous = _state.field(field);
      _setField(
        field,
        previous.copyWith(
          suggestions: suggestions,
          searching: false,
          searched: true,
          clearError: true,
        ),
      );
    } on PlaceSearchCancelled {
      // A newer query in the same session superseded this request.
    } catch (error) {
      if (!_isCurrent(field, revision)) return;
      final previous = _state.field(field);
      _setField(
        field,
        previous.copyWith(
          suggestions: const [],
          searching: false,
          searched: true,
          error: _message(error),
        ),
      );
    }
  }

  Future<GeoPoint?> _loadBias() {
    return _biasFuture ??= () async {
      try {
        return await _deviceLocation.current();
      } catch (_) {
        return null;
      }
    }();
  }

  void _applyPoint(RiderPlaceField field, GeoPoint point) {
    if (field == RiderPlaceField.pickup) {
      _rider.setPickup(point);
    } else {
      _rider.setDestination(point);
    }
  }

  int _nextRevision(RiderPlaceField field) {
    final next = (_revisions[field] ?? 0) + 1;
    _revisions[field] = next;
    return next;
  }

  bool _isCurrent(RiderPlaceField field, int revision) =>
      !_disposed && _revisions[field] == revision;

  void _endSession(RiderPlaceField field) {
    final token = _sessionTokens.remove(field);
    if (token != null) _repository.cancelSession(token);
  }

  void _setField(RiderPlaceField field, PlaceFieldSearchState value) {
    if (_disposed) return;
    _state = _state.replace(field, value);
    notifyListeners();
  }

  String _message(Object error) =>
      error.toString().replaceFirst('Exception: ', '');

  String _newSessionToken() {
    final random = Random.secure();
    final bytes = List<int>.generate(16, (_) => random.nextInt(256));
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    final hex = bytes
        .map((value) => value.toRadixString(16).padLeft(2, '0'))
        .join();
    return '${hex.substring(0, 8)}-'
        '${hex.substring(8, 12)}-'
        '${hex.substring(12, 16)}-'
        '${hex.substring(16, 20)}-'
        '${hex.substring(20)}';
  }

  @override
  void dispose() {
    _disposed = true;
    for (final timer in _debounces.values) {
      timer.cancel();
    }
    for (final token in _sessionTokens.values) {
      _repository.cancelSession(token);
    }
    super.dispose();
  }
}
