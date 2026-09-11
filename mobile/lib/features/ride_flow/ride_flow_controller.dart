import 'dart:async';
import 'package:flutter/foundation.dart';
import '../../core/network/api_exception.dart';
import 'ride_flow_repository.dart';

/// Server-owned state. Polling is serialized with commands and stopped in background.
class RideFlowController extends ChangeNotifier {
  RideFlowController(this.repository, {this.rideId, this.interval = const Duration(seconds: 5)}) {
    refresh();
  }
  final RideFlowRepository repository;
  final String? rideId;
  final Duration interval;
  bool get driver => rideId == null;
  Json? current;
  Json? location;
  List<Json> offers = [];
  List<Json> requests = [];
  List<Json> history = [];
  bool busy = false;
  bool loaded = false;
  String? error;
  bool _disposed = false;
  bool _foreground = true;
  Timer? _timer;
  String? get status {
    if (driver) return current?['status'] as String?;
    return (current?['trip']?['status'] ?? current?['status']) as String?;
  }

  Future<Json?> _optional(String path) async {
    try { return await repository.get(path); }
    on ApiException catch (e) { if (e.statusCode == 404) return null; rethrow; }
  }
  List<Json> _list(Json json, String key) =>
    (json[key] as List? ?? []).map((e) => Map<String, dynamic>.from(e as Map)).toList();

  Future<void> _load() async {
    if (driver) {
      final trip = await _optional('/v1/driver/trip');
      final feed = trip == null ? await repository.get('/v1/driver/marketplace/ride-requests') : <String,dynamic>{};
      final past = await repository.get('/v1/driver/trips');
      if (_disposed) return;
      current = trip;
      requests = _list(feed, 'ride_requests');
      history = _list(past, 'trips');
    } else {
      final ride = await repository.get('/v1/ride-requests/$rideId');
      final active = ride['trip'] != null && !['completed','cancelled'].contains(ride['trip']['status']);
      final comparison = ride['trip'] == null && ride['status'] == 'requested'
        ? await repository.get('/v1/ride-requests/$rideId/offers') : <String,dynamic>{};
      final position = active ? await _optional('/v1/ride-requests/$rideId/driver-location') : null;
      if (_disposed) return;
      current = ride;
      offers = _list(comparison, 'offers');
      location = position;
    }
    loaded = true;
  }

  Future<void> refresh() => _run(_load);

  Future<void> act(String path, {Json? data, bool put = false}) => _run(() async {
    try { await repository.act(path, data: data, put: put); }
    catch (_) {
      // A response may be lost after commit, or the offer may have changed.
      // Reload authoritative state without claiming that a failed command succeeded.
      try { await _load(); } catch (_) {}
      rethrow;
    }
    await _load();
  });

  Future<void> _run(Future<void> Function() task) async {
    if (busy || _disposed || !_foreground) return;
    _timer?.cancel();
    busy = true;
    error = null;
    notifyListeners();
    try { await task(); } catch (e) { error = '$e'; }
    finally {
      busy = false;
      if (!_disposed) { notifyListeners(); _schedule(); }
    }
  }
  void _schedule() {
    _timer?.cancel();
    if (!_disposed && _foreground) _timer = Timer(interval, refresh);
  }
  void setForeground(bool value) {
    _foreground = value;
    _timer?.cancel();
    if (value) refresh();
  }
  @override
  void dispose() { _disposed = true; _timer?.cancel(); super.dispose(); }
}
