import 'package:shared_preferences/shared_preferences.dart';

class LastMapLocation {
  const LastMapLocation({required this.latitude, required this.longitude});

  final double latitude;
  final double longitude;

  bool get isValid =>
      latitude.isFinite &&
      longitude.isFinite &&
      latitude >= -90 &&
      latitude <= 90 &&
      longitude >= -180 &&
      longitude <= 180;
}

abstract interface class LastMapLocationStore {
  Future<LastMapLocation?> read();
  Future<void> save(LastMapLocation location);
}

class PreferencesLastMapLocationStore implements LastMapLocationStore {
  static const _latitudeKey = 'last_map_location_latitude';
  static const _longitudeKey = 'last_map_location_longitude';

  @override
  Future<LastMapLocation?> read() async {
    try {
      final preferences = await SharedPreferences.getInstance();
      final latitude = preferences.getDouble(_latitudeKey);
      final longitude = preferences.getDouble(_longitudeKey);
      if (latitude == null || longitude == null) return null;
      final location = LastMapLocation(
        latitude: latitude,
        longitude: longitude,
      );
      return location.isValid ? location : null;
    } catch (_) {
      return null;
    }
  }

  @override
  Future<void> save(LastMapLocation location) async {
    if (!location.isValid) return;
    try {
      final preferences = await SharedPreferences.getInstance();
      await preferences.setDouble(_latitudeKey, location.latitude);
      await preferences.setDouble(_longitudeKey, location.longitude);
    } catch (_) {
      // A cache failure must never block the live location flow.
    }
  }
}
