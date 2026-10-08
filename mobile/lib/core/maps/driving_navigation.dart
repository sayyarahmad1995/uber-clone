import 'package:flutter/services.dart';

import 'ride_map.dart';

abstract interface class DrivingNavigation {
  Future<bool> open(RideMapPoint destination);
}

/// Android owns app selection and falls back to the Google Maps website.
class AndroidDrivingNavigation implements DrivingNavigation {
  const AndroidDrivingNavigation();

  static const _channel = MethodChannel('higo/navigation');

  @override
  Future<bool> open(RideMapPoint destination) async {
    if (!destination.latitude.isFinite ||
        !destination.longitude.isFinite ||
        destination.latitude.abs() > 90 ||
        destination.longitude.abs() > 180) {
      return false;
    }
    try {
      return await _channel.invokeMethod<bool>('openGoogleMaps', {
            'latitude': destination.latitude,
            'longitude': destination.longitude,
          }) ==
          true;
    } on PlatformException {
      return false;
    } on MissingPluginException {
      return false;
    }
  }
}
