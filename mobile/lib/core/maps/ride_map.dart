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
class RideMap extends StatelessWidget {
  const RideMap({
    super.key,
    required this.tiles,
    this.markers = const [],
    this.initialCenter = const LatLng(24.8607, 67.0011),
    this.initialZoom = 12,
    this.onTap,
  });

  final MapTiles tiles;
  final List<RideMapMarker> markers;
  final LatLng initialCenter;
  final double initialZoom;
  final ValueChanged<LatLng>? onTap;

  @override
  Widget build(BuildContext context) => FlutterMap(
    options: MapOptions(
      initialCenter: initialCenter,
      initialZoom: initialZoom,
      onTap: onTap == null ? null : (_, point) => onTap!(point),
    ),
    children: [
      TileLayer(
        urlTemplate: tiles.urlTemplate,
        userAgentPackageName: tiles.userAgentPackageName,
      ),
      if (markers.isNotEmpty)
        MarkerLayer(
          markers: markers
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
        attributions: [TextSourceAttribution(tiles.attribution)],
      ),
    ],
  );
}

class _MapMarker extends StatelessWidget {
  const _MapMarker({required this.marker});

  final RideMapMarker marker;

  @override
  Widget build(BuildContext context) => Tooltip(
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
