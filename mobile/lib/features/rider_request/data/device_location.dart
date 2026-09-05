import 'package:geolocator/geolocator.dart';

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
