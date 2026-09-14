import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';

import '../theme/app_theme.dart';
import 'map_tiles.dart';

class RideMapMarker {
  const RideMapMarker({
    required this.point,
    required this.icon,
    required this.color,
    this.label,
  });

  final LatLng point;
  final IconData icon;
  final Color color;
  final String? label;
}

/// Reusable map layer for Rider and Driver dashboards.
///
/// Feature screens decide what the markers mean; this widget only owns the map
/// rendering contract and keeps tile attribution consistent across the client.
class RideMap extends StatefulWidget {
  const RideMap({
    super.key,
    required this.tiles,
    this.mapController,
    this.markers = const [],
    this.initialCenter = const LatLng(24.8607, 67.0011),
    this.initialZoom = 12,
    this.onTap,
  });

  final MapTiles tiles;
  final MapController? mapController;
  final List<RideMapMarker> markers;
  final LatLng initialCenter;
  final double initialZoom;
  final ValueChanged<LatLng>? onTap;

  @override
  State<RideMap> createState() => _RideMapState();
}

class _RideMapState extends State<RideMap> {
  static const _autoCenterLabel = 'Your published location';

  String? _lastAutoCenteredPoint;

  @override
  void initState() {
    super.initState();
    _scheduleAutoCenter();
  }

  @override
  void didUpdateWidget(covariant RideMap oldWidget) {
    super.didUpdateWidget(oldWidget);
    _scheduleAutoCenter();
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
    final controller = widget.mapController;
    if (controller == null) return;

    final marker = _autoCenterMarker();
    if (marker == null) return;

    final key = '${marker.point.latitude},${marker.point.longitude}';
    if (_lastAutoCenteredPoint == key) return;
    _lastAutoCenteredPoint = key;

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      controller.move(marker.point, 15);
    });
  }

  @override
  Widget build(BuildContext context) {
    return FlutterMap(
      mapController: widget.mapController,
      options: MapOptions(
        initialCenter: widget.initialCenter,
        initialZoom: widget.initialZoom,
        onTap: widget.onTap == null ? null : (_, point) => widget.onTap!(point),
      ),
      children: [
        TileLayer(
          urlTemplate: widget.tiles.urlTemplate,
          userAgentPackageName: widget.tiles.userAgentPackageName,
        ),
        if (widget.markers.isNotEmpty)
          MarkerLayer(
            markers: widget.markers
                .map(
                  (marker) => Marker(
                    point: marker.point,
                    width: 56,
                    height: 56,
                    child: _MapMarker(marker: marker),
                  ),
                )
                .toList(growable: false),
          ),
        RichAttributionWidget(
          attributions: [TextSourceAttribution(widget.tiles.attribution)],
        ),
      ],
    );
  }
}

class _MapMarker extends StatelessWidget {
  const _MapMarker({required this.marker});

  final RideMapMarker marker;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: marker.label ?? '',
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: const BorderRadius.all(Radius.circular(AppRadii.lg)),
          boxShadow: const [
            BoxShadow(
              color: Colors.black26,
              blurRadius: 8,
              offset: Offset(0, 2),
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xs),
          child: Icon(marker.icon, color: marker.color, size: 32),
        ),
      ),
    );
  }
}
