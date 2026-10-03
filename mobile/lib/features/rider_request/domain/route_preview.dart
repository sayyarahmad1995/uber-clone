import 'ride_request.dart';

class RoutePreview {
  RoutePreview({required List<RouteOption> routes})
    : routes = List.unmodifiable(routes) {
    if (routes.isEmpty) {
      throw const FormatException('Route preview has no routes.');
    }
    final ids = <String>{};
    var recommendedCount = 0;
    for (final route in routes) {
      if (!ids.add(route.id)) {
        throw const FormatException('Route preview contains duplicate IDs.');
      }
      if (route.recommended) {
        recommendedCount++;
      }
    }
    if (recommendedCount != 1) {
      throw const FormatException(
        'Route preview must contain exactly one recommended route.',
      );
    }
  }

  final List<RouteOption> routes;

  RouteOption get recommended =>
      routes.firstWhere((route) => route.recommended);

  RouteOption? routeById(String? id) {
    if (id == null) return null;
    for (final route in routes) {
      if (route.id == id) return route;
    }
    return null;
  }

  factory RoutePreview.fromJson(Map<String, dynamic> json) {
    final rawRoutes = json['routes'] as List<dynamic>?;
    if (rawRoutes == null) {
      throw const FormatException('Route preview routes are missing.');
    }
    return RoutePreview(
      routes: rawRoutes
          .map((route) => RouteOption.fromJson(route as Map<String, dynamic>))
          .toList(growable: false),
    );
  }
}

class RouteOption {
  RouteOption({
    required this.id,
    required this.recommended,
    required this.distanceMeters,
    required this.durationSeconds,
    required this.encodedPolyline,
  }) : points = decodeEncodedPolyline(encodedPolyline) {
    if (id.trim().isEmpty || distanceMeters <= 0 || durationSeconds <= 0) {
      throw const FormatException('Route option is invalid.');
    }
  }

  final String id;
  final bool recommended;
  final int distanceMeters;
  final int durationSeconds;
  final String encodedPolyline;
  final List<GeoPoint> points;

  factory RouteOption.fromJson(Map<String, dynamic> json) {
    return RouteOption(
      id: json['id'] as String,
      recommended: json['recommended'] as bool,
      distanceMeters: (json['distance_meters'] as num).toInt(),
      durationSeconds: (json['duration_seconds'] as num).toInt(),
      encodedPolyline: json['encoded_polyline'] as String,
    );
  }
}

List<GeoPoint> decodeEncodedPolyline(String encoded) {
  if (encoded.isEmpty) {
    throw const FormatException('Route polyline is empty.');
  }

  final points = <GeoPoint>[];
  var latitude = 0;
  var longitude = 0;
  var index = 0;

  while (index < encoded.length) {
    final lat = _decodeValue(encoded, index);
    index = lat.$2;
    latitude += lat.$1;

    if (index >= encoded.length) {
      throw const FormatException('Route polyline is truncated.');
    }
    final lng = _decodeValue(encoded, index);
    index = lng.$2;
    longitude += lng.$1;

    points.add(GeoPoint(latitude: latitude / 1e5, longitude: longitude / 1e5));
  }

  if (points.length < 2) {
    throw const FormatException('Route polyline has too few points.');
  }
  return List.unmodifiable(points);
}

(int, int) _decodeValue(String encoded, int start) {
  var result = 0;
  var shift = 0;
  var index = start;

  while (true) {
    if (index >= encoded.length) {
      throw const FormatException('Route polyline is truncated.');
    }
    final value = encoded.codeUnitAt(index++) - 63;
    if (value < 0) {
      throw const FormatException('Route polyline contains invalid data.');
    }
    result |= (value & 0x1f) << shift;
    shift += 5;
    if (value < 0x20) {
      break;
    }
    if (shift > 30) {
      throw const FormatException('Route polyline value is invalid.');
    }
  }

  final delta = (result & 1) != 0 ? ~(result >> 1) : result >> 1;
  return (delta, index);
}
