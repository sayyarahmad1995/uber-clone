import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uber_clone/core/maps/last_map_location_store.dart';
import 'package:uber_clone/features/rider_request/data/device_location.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  test('persists and restores a valid last map location', () async {
    final store = PreferencesLastMapLocationStore();

    await store.save(
      const LastMapLocation(latitude: 33.64984, longitude: 72.97685),
    );

    final restored = await store.read();

    expect(restored, isNotNull);
    expect(restored!.latitude, 33.64984);
    expect(restored.longitude, 72.97685);
  });

  test('ignores invalid coordinates instead of caching them', () async {
    final store = PreferencesLastMapLocationStore();

    await store.save(
      const LastMapLocation(latitude: 120, longitude: 72.97685),
    );

    expect(await store.read(), isNull);
  });

  test('successful device location reads update the last map cache', () async {
    final store = PreferencesLastMapLocationStore();
    final location = CachingDeviceLocation(
      const StaticDeviceLocation(
        GeoPoint(latitude: 33.64984, longitude: 72.97685),
      ),
      store,
    );

    final point = await location.current();
    final cached = await store.read();

    expect(point.latitude, 33.64984);
    expect(cached, isNotNull);
    expect(cached!.latitude, 33.64984);
    expect(cached.longitude, 72.97685);
  });

  test('cache write failures do not block live location reads', () async {
    final location = CachingDeviceLocation(
      const StaticDeviceLocation(
        GeoPoint(latitude: 33.64984, longitude: 72.97685),
      ),
      const FailingLastMapLocationStore(),
    );

    final point = await location.current();

    expect(point.latitude, 33.64984);
    expect(point.longitude, 72.97685);
  });
}

class StaticDeviceLocation implements DeviceLocation {
  const StaticDeviceLocation(this.point);

  final GeoPoint point;

  @override
  Future<GeoPoint> current() async => point;
}

class FailingLastMapLocationStore implements LastMapLocationStore {
  const FailingLastMapLocationStore();

  @override
  Future<LastMapLocation?> read() async => null;

  @override
  Future<void> save(LastMapLocation location) async {
    throw Exception('Cache unavailable');
  }
}
