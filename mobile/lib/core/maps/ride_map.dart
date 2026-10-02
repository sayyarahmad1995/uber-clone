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
  Future<void> Function(List<RideMapPoint> points, double padding)? _fit;
  (RideMapPoint point, double zoom)? _pending;
  (List<RideMapPoint> points, double padding)? _pendingFit;
  RideMapPoint? _center;

  RideMapPoint? get center => _center;

  Future<void> move(RideMapPoint point, double zoom) async {
    if (!point.isValid) return;
    final move = _move;
    if (move == null) {
      _pending = (point, zoom);
      return;
    }
    await move(point, zoom);
  }

  Future<void> fit(
    Iterable<RideMapPoint> points, {
    double padding = 48,
  }) async {
    final valid = points\n        .where((point) => point.isValid)\n        .toList(growable: false);
    if (valid.isEmpty) return;
    final fit = _fit;
    if (fit == null) {
      _pendingFit = (valid, padding);
      return;
    }
    await fit(valid, padding);
  }

  void _attach(
    Future<void> Function(RideMapPoint point, double zoom) move,
    Future<void> Function(List<RideMapPoint> points, double padding) fit,
  ) {
    _move = move;
    _fit = fit;

    final pending = _pending;
    _pending = null;
    if (pending != null) {
      move(pending.$1, pending.$2);
    }

    final pendingFit = _pendingFit;
    _pendingFit = null;
    if (pendingFit != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        fit(pendingFit.$1, pendingFit.$2);
      });
    }
  }

  void _updateCenter(RideMapPoint point) {
    if (point.isValid) _center = point;
  }

  void _detach() {
    _move = null;
    _fit = null;
  }
}

class RideMapMarker {
  const RideMapMarker({
    required this.point,
    required this.color,
    this.label,
    this.onDragEnd,
  });

  final RideMapPoint point;
  final Color color;
  final String? label;
  final ValueChanged<RideMapPoint>? onDragEnd;
}

class RideMapPolyline {
  const RideMapPolyline({
    required this.points,
    required this.color,
    this.width = 5,
  });

  final List<RideMapPoint> points;
  final Color color;
  final int width;
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
    this.polylines = const [],
    this.padding = EdgeInsets.zero,
    this.initialCenter = defaultCenter,
    this.initialZoom = 12,
    this.onTap,
    this.showCenterPin = false,
    this.centerPinColor,
  });

  static const defaultCenter = RideMapPoint(24.8607, 67.0011);

  final RideMapController? mapController;
  final List<RideMapMarker> markers;
  final List<RideMapPolyline> polylines;
  final EdgeInsets padding;
  final RideMapPoint initialCenter;
  final double initialZoom;
  final ValueChanged<RideMapPoint>? onTap;
  final bool showCenterPin;
  final Color? centerPinColor;

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
  late RideMapPoint _cameraCenter;

  @override
  void initState() {
    super.initState();
    _cameraCenter = widget.initialCenter;
    widget.mapController?._updateCenter(_cameraCenter);
    _restoreCachedCenter();
    _scheduleAutoCenter();
  }

  @override
  void didUpdateWidget(covariant RideMap oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.mapController != widget.mapController) {
      oldWidget.mapController?._detach();
      if (_googleController != null) {
        widget.mapController?._attach(_explicitMove, _fitPoints);
        widget.mapController?._updateCenter(_cameraCenter);
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
    widget.mapController?._attach(_explicitMove, _fitPoints);
    _scheduleAutoCenter();
  }

  Future<void> _explicitMove(RideMapPoint point, double zoom) async {
    _userMovedCamera = false;
    await _moveCamera(point, zoom);
  }

  Future<void> _moveCamera(RideMapPoint point, double zoom) async {
    final controller = _googleController;
    if (controller == null || !point.isValid) return;
    _cameraCenter = point;
    widget.mapController?._updateCenter(point);
    _programmaticCameraMove = true;
    try {
      await controller.moveCamera(
        google.CameraUpdate.newLatLngZoom(_googlePoint(point), zoom),
      );
    } finally {
      _programmaticCameraMove = false;
    }
  }

  Future<void> _fitPoints(List<RideMapPoint> points, double padding) async {
    final controller = _googleController;
    final valid = points.where((point) => point.isValid).toList(growable: false);
    if (controller == null || valid.isEmpty) return;
    if (valid.length == 1) {
      await _moveCamera(valid.single, 16);
      return;
    }

    var minLatitude = valid.first.latitude;
    var maxLatitude = valid.first.latitude;
    var minLongitude = valid.first.longitude;
    var maxLongitude = valid.first.longitude;
    for (final point in valid.skip(1)) {
      if (point.latitude < minLatitude) minLatitude = point.latitude;
      if (point.latitude > maxLatitude) maxLatitude = point.latitude;
      if (point.longitude < minLongitude) minLongitude = point.longitude;
      if (point.longitude > maxLongitude) maxLongitude = point.longitude;
    }
    if (minLatitude == maxLatitude && minLongitude == maxLongitude) {
      await _moveCamera(valid.first, 16);
      return;
    }

    _programmaticCameraMove = true;
    try {
      await controller.animateCamera(
        google.CameraUpdate.newLatLngBounds(
          google.LatLngBounds(
            southwest: google.LatLng(minLatitude, minLongitude),
            northeast: google.LatLng(maxLatitude, maxLongitude),
          ),
          padding,
        ),
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

  void _onCameraMove(google.CameraPosition position) {
    final point = RideMapPoint(
      position.target.latitude,
      position.target.longitude,
    );
    _cameraCenter = point;
    widget.mapController?._updateCenter(point);
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
    return Stack(
      fit: StackFit.expand,
      children: [
        google.GoogleMap(
          initialCameraPosition: google.CameraPosition(
            target: _googlePoint(widget.initialCenter),
            zoom: widget.initialZoom,
          ),
          onMapCreated: _onMapCreated,
          onTap: widget.onTap == null
              ? null
              : (point) => widget.onTap!(
                  RideMapPoint(point.latitude, point.longitude),
                ),
          onCameraMoveStarted: _onCameraMoveStarted,
          onCameraMove: _onCameraMove,
          onCameraIdle: _onCameraIdle,
          padding: widget.padding,
          markers: {
            for (final entry in widget.markers.indexed)
              if (entry.$2.point.isValid) _googleMarker(entry.$1, entry.$2),
          },
          polylines: {
            for (final entry in widget.polylines.indexed)
              if (entry.$2.points
                      .where((point) => point.isValid)
                      .length >=
                  2)
                _googlePolyline(entry.$1, entry.$2),
          },
          myLocationButtonEnabled: false,
          zoomControlsEnabled: false,
          mapToolbarEnabled: false,
          compassEnabled: false,
        ),
        if (widget.showCenterPin)
          IgnorePointer(
            child: Center(
              child: Transform.translate(
                offset: const Offset(0, -24),
                child: Icon(
                  Icons.location_pin,
                  key: const Key('rideMapCenterPin'),
                  size: 48,
                  color:
                      widget.centerPinColor ??
                      Theme.of(context).colorScheme.primary,
                ),
              ),
            ),
          ),
      ],
    );
  }

  google.Marker _googleMarker(int index, RideMapMarker marker) {
    return google.Marker(
      markerId: google.MarkerId(
        '${marker.label ?? 'marker'}-$index-'
        '${marker.point.latitude}-${marker.point.longitude}',
      ),
      position: _googlePoint(marker.point),
      draggable: marker.onDragEnd != null,
      onDragEnd: marker.onDragEnd == null
          ? null
          : (point) => marker.onDragEnd!(
              RideMapPoint(point.latitude, point.longitude),
            ),
      icon: google.BitmapDescriptor.defaultMarkerWithHue(
        HSVColor.fromColor(marker.color).hue,
      ),
      infoWindow: marker.label == null
          ? google.InfoWindow.noText
          : google.InfoWindow(title: marker.label),
    );
  }

  google.Polyline _googlePolyline(int index, RideMapPolyline polyline) {
    return google.Polyline(
      polylineId: google.PolylineId('route-$index'),
      points: polyline.points
          .where((point) => point.isValid)
          .map(_googlePoint)
          .toList(growable: false),
      color: polyline.color,
      width: polyline.width,
    );
  }

  google.LatLng _googlePoint(RideMapPoint point) =>
      google.LatLng(point.latitude, point.longitude);
}
