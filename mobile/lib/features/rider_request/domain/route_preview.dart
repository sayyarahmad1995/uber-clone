import 'ride_request.dart';

class RoutePreview {
  RoutePreview({
    required this.distanceMeters,
    required this.durationSeconds,
    required this.encodedPolyline,
  }) : points = decodeEncodedPolyline(encodedPolyline);

  final int distanceMeters;
  final int durationSeconds;
  final String encodedPolyline;
  final List<GeoPoint> points;

  factory RoutePreview.fromJson(Map<String, dynamic> json) {
    final route = json['route'] as Map<String, dynamic>;
    return RoutePreview(
      distanceMeters: (route['distance_meters'] as num).toInt(),
      durationSeconds: (route['duration_seconds'] as num).toInt(),
      encodedPolyline: route['encoded_polyline'] as String,
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
