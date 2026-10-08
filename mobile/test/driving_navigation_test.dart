import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/maps/driving_navigation.dart';
import 'package:uber_clone/core/maps/ride_map.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  const channel = MethodChannel('higo/navigation');
  const navigation = AndroidDrivingNavigation();

  tearDown(() {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, null);
  });

  test('navigation passes exact coordinates to the Android boundary', () async {
    MethodCall? received;
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (call) async {
          received = call;
          return true;
        });
    expect(await navigation.open(const RideMapPoint(-24.87, 67.02)), isTrue);
    expect(received?.method, 'openGoogleMaps');
    expect(received?.arguments, {
      'latitude': -24.87,
      'longitude': 67.02,
    });
  });

  test('invalid coordinates cannot reach the platform launcher', () async {
    var calls = 0;
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (_) async {
          calls++;
          return true;
        });
    for (final point in [
      const RideMapPoint(91, 67),
      const RideMapPoint(24, -181),
      const RideMapPoint(double.nan, 67),
      const RideMapPoint(24, double.infinity),
    ]) {
      expect(await navigation.open(point), isFalse);
    }
    expect(calls, 0);
  });

  test('platform rejection is returned as a recoverable failure', () async {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (_) async => false);
    expect(await navigation.open(const RideMapPoint(24, 67)), isFalse);
  });

  test('platform exception is returned as a recoverable failure', () async {
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (_) async {
          throw PlatformException(code: 'unavailable');
        });
    expect(await navigation.open(const RideMapPoint(24, 67)), isFalse);
  });

  test('unsupported platform is returned as a recoverable failure', () async {
    expect(await navigation.open(const RideMapPoint(24, 67)), isFalse);
  });
}
