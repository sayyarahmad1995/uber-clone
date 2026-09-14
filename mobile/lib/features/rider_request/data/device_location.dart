import 'package:geolocator/geolocator.dart';

import '../../../core/maps/last_map_location_store.dart';
import '../domain/ride_request.dart';

abstract interface class DeviceLocation {
  Future<GeoPoint> current();
}

class LocationUnavailable implements Exception {
  const LocationUnavailable(this.message);
  final String message;
  @override
  String toString() => message;
}

class CachingDeviceLocation implements DeviceLocation {
  CachingDeviceLocation(this._delegate, this._cache);

  final DeviceLocation _delegate;
  final LastMapLocationStore _cache;

  @override
  Future<GeoPoint> current() async {
    final point = await _delegate.current();
    try {
      await _cache.save(
        LastMapLocation(
          latitude: point.latitude,
          longitude: point.longitude,
        ),
      );
    } catch (_) {
      // A cache write must never block a live location read.
    }
    return point;
  }
}

class GeolocatorDeviceLocation implements DeviceLocation {
  @override
  Future<GeoPoint> current() async {
    if (!await Geolocator.isLocationServiceEnabled()) {
      throw const LocationUnavailable(
        'Turn on device location to use your current position.',
      );
    }
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever) {
      throw const LocationUnavailable(
        'Location permission is required to use your current position.',
      );
    }
    final position = await Geolocator.getCurrentPosition();
    return GeoPoint(latitude: position.latitude, longitude: position.longitude);
  }
}
