import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart' as google;

import 'last_map_location_store.dart';

@immutable
class RideMapPoint {
  const RideMapPoint(this.latitude, this.longitude);

  final double latitude;
  final double longitude;

  bool get isValid =>
      latitude.isFinite &&
      longitude.isFinite &&
      latitude >= -90 &&
      latitude <= 90 &&
      longitude >= -180 &&
      longitude <= 180;

  @override
  bool operator ==(Object other) =>
      other is RideMapPoint &&
      other.latitude == latitude &&
      other.longitude == longitude;

  @override
  int get hashCode => Object.hash(latitude, longitude);
}

/// Provider-neutral camera controller used by Rider and Driver feature screens.
///
/// Explicit feature-level camera moves (for example, pressing the focus button)
/// clear any prior user-pan override. Automatic marker updates do not.
class RideMapController {
  Future<void> Function(RideMapPoint point, double zoom)? _move;
  (RideMapPoint point, double zoom)? _pending;

  Future<void> move(RideMapPoint point, double zoom) async {
    if (!point.isValid) return;
    final move = _move;
    if (move == null) {
      _pending = (point, zoom);
      return;
    }
    await move(point, zoom);
  }

  void _attach(Future<void> Function(RideMapPoint point, double zoom) move) {
    _move = move;
    final pending = _pending;
    _pending = null;
    if (pending != null) {
      move(pending.$1, pending.$2);
    }
  }

  void _detach() {
    _move = null;
  }
}

class RideMapMarker {
  const RideMapMarker({required this.point, required this.color, this.label});

  final RideMapPoint point;
  final Color color;
  final String? label;
}

/// Shared Google Maps rendering boundary for Rider and Driver dashboards.
///
/// Feature screens own marker meaning and booking behavior. Google-specific
/// types remain inside this file so routing/map provider details do not leak
/// into feature or domain contracts.
class RideMap extends StatefulWidget {
  const RideMap({
    super.key,
    this.mapController,
    this.markers = const [],
    this.initialCenter = defaultCenter,
    this.initialZoom = 12,
    this.onTap,
  });

  static const defaultCenter = RideMapPoint(24.8607, 67.0011);

  final RideMapController? mapController;
  final List<RideMapMarker> markers;
  final RideMapPoint initialCenter;
  final double initialZoom;
  final ValueChanged<RideMapPoint>? onTap;

  @override
  State<RideMap> createState() => _RideMapState();
}

class _RideMapState extends State<RideMap> {
  static const _autoCenterLabel = 'Your published location';

  google.GoogleMapController? _googleController;
  String? _lastAutoCenteredPoint;
  bool _restoredCachedCenter = false;
  bool _programmaticCameraMove = false;
  bool _userMovedCamera = false;

  @override
  void initState() {
    super.initState();
    _restoreCachedCenter();
    _scheduleAutoCenter();
  }

  @override
  void didUpdateWidget(covariant RideMap oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.mapController != widget.mapController) {
      oldWidget.mapController?._detach();
      if (_googleController != null) {
        widget.mapController?._attach(_explicitMove);
      }
    }
    _scheduleAutoCenter();
  }

  @override
  void dispose() {
    widget.mapController?._detach();
    super.dispose();
  }

  void _onMapCreated(google.GoogleMapController controller) {
    _googleController = controller;
    widget.mapController?._attach(_explicitMove);
    _scheduleAutoCenter();
  }

  Future<void> _explicitMove(RideMapPoint point, double zoom) async {
    _userMovedCamera = false;
    await _moveCamera(point, zoom);
  }

  Future<void> _moveCamera(RideMapPoint point, double zoom) async {
    final controller = _googleController;
    if (controller == null || !point.isValid) return;
    _programmaticCameraMove = true;
    try {
      await controller.moveCamera(
        google.CameraUpdate.newLatLngZoom(_googlePoint(point), zoom),
      );
    } finally {
      _programmaticCameraMove = false;
    }
  }

  void _onCameraMoveStarted() {
    if (!_programmaticCameraMove) {
      _userMovedCamera = true;
    }
  }

  void _onCameraIdle() {
    _programmaticCameraMove = false;
  }

  Future<void> _restoreCachedCenter() async {
    if (_restoredCachedCenter) return;
    final location = await PreferencesLastMapLocationStore().read();
    if (!mounted || location == null || _restoredCachedCenter) return;
    _restoredCachedCenter = true;
    final center = RideMapPoint(location.latitude, location.longitude);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || _userMovedCamera) return;
      _moveCamera(center, 15);
    });
  }

  RideMapMarker? _autoCenterMarker() {
    for (final marker in widget.markers) {
      if (marker.label == _autoCenterLabel) {
        return marker;
      }
    }
    return null;
  }

  void _scheduleAutoCenter() {
    if (_googleController == null || _userMovedCamera) return;
    final marker = _autoCenterMarker();
    if (marker == null || !marker.point.isValid) return;

    final key = '${marker.point.latitude},${marker.point.longitude}';
    if (_lastAutoCenteredPoint == key) return;
    _lastAutoCenteredPoint = key;

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || _userMovedCamera) return;
      _moveCamera(marker.point, 15);
    });
  }

  @override
  Widget build(BuildContext context) {
    return google.GoogleMap(
      initialCameraPosition: google.CameraPosition(
        target: _googlePoint(widget.initialCenter),
        zoom: widget.initialZoom,
      ),
      onMapCreated: _onMapCreated,
      onTap: widget.onTap == null
          ? null
          : (point) =>
                widget.onTap!(RideMapPoint(point.latitude, point.longitude)),
      onCameraMoveStarted: _onCameraMoveStarted,
      onCameraIdle: _onCameraIdle,
      markers: {
        for (final entry in widget.markers.indexed)
          if (entry.$2.point.isValid) _googleMarker(entry.$1, entry.$2),
      },
      myLocationButtonEnabled: false,
      zoomControlsEnabled: false,
      mapToolbarEnabled: false,
      compassEnabled: false,
    );
  }

  google.Marker _googleMarker(int index, RideMapMarker marker) {
    return google.Marker(
      markerId: google.MarkerId(
        '${marker.label ?? 'marker'}-$index-'
        '${marker.point.latitude}-${marker.point.longitude}',
      ),
      position: _googlePoint(marker.point),
      icon: google.BitmapDescriptor.defaultMarkerWithHue(
        HSVColor.fromColor(marker.color).hue,
      ),
      infoWindow: marker.label == null
          ? google.InfoWindow.noText
          : google.InfoWindow(title: marker.label),
    );
  }

  google.LatLng _googlePoint(RideMapPoint point) =>
      google.LatLng(point.latitude, point.longitude);
}
