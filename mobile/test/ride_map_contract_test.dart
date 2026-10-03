import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/maps/ride_map.dart';

void main() {
  test('RideMapPoint validates geographic coordinate bounds', () {
    expect(const RideMapPoint(24.8607, 67.0011).isValid, isTrue);
    expect(const RideMapPoint(-90, -180).isValid, isTrue);
    expect(const RideMapPoint(90, 180).isValid, isTrue);
    expect(const RideMapPoint(91, 0).isValid, isFalse);
    expect(const RideMapPoint(0, 181).isValid, isFalse);
    expect(const RideMapPoint(double.nan, 0).isValid, isFalse);
  });

  test('RideMap exposes provider-neutral center-pin selection state', () {
    final source = File('lib/core/maps/ride_map.dart').readAsStringSync();

    expect(source, contains('RideMapPoint? get center'));
    expect(source, contains('showCenterPin'));
    expect(source, contains("Key('rideMapCenterPin')"));
    expect(source, contains('onCameraMove: _onCameraMove'));
  });

  test('Rider map selection exposes explicit pin-mode controls', () {
    final source = File(
      'lib/features/rider_request/presentation/rider_request_screen.dart',
    ).readAsStringSync();

    expect(source, contains("'pickupSetOnMapButton'"));
    expect(source, contains("'destinationSetOnMapButton'"));
    expect(source, contains("Key('confirmPinButton')"));
    expect(source, contains('_confirmPinSelection'));
    expect(source, contains('FloatingActionButton.extended'));
    expect(
      source,
      contains('showCenterPin: active == null && _pinSelectionMode'),
    );
  });

  test('pin mode and dashboard extent do not reposition the Rider map', () {
    final source = File(
      'lib/features/rider_request/presentation/rider_request_screen.dart',
    ).readAsStringSync();

    expect(source, isNot(contains('DashboardPanelSessionScope')));
    expect(source, isNot(contains('mapBottomPadding')));
    expect(source, isNot(contains('panelExpanded')));
    expect(
      source,
      contains("preview.routes.map((route) => route.encodedPolyline).join('|')"),
    );
  });

  test('RideMap translates Google POI taps behind provider boundary', () {
    final source = File('lib/core/maps/ride_map.dart').readAsStringSync();
    final rider = File(
      'lib/features/rider_request/presentation/rider_request_screen.dart',
    ).readAsStringSync();

    expect(source, contains('class RideMapPlace'));
    expect(source, contains('onPointOfInterestTap'));
    expect(source, contains('PointOfInterestTapEvent'));
    expect(source, contains('RideMapPlace(placeId: placeId)'));
    expect(rider, contains('onPlaceTap: active == null ? _handlePlaceTap : null'));
    expect(rider, contains('selectPlaceId(field, place.placeId)'));
  });

  test('RideMap exposes provider-neutral route rendering and camera fit', () {
    final source = File('lib/core/maps/ride_map.dart').readAsStringSync();
    final rider = File(
      'lib/features/rider_request/presentation/rider_request_screen.dart',
    ).readAsStringSync();

    expect(source, contains('class RideMapPolyline'));
    expect(source, contains('consumeTapEvents: polyline.onTap != null'));
    expect(source, contains('onTap: polyline.onTap'));
    expect(source, contains('Future<void> fit('));
    expect(source, contains('newLatLngBounds'));
    expect(source, contains('polylines: {'));
    expect(source, contains('padding: widget.padding'));
    expect(rider, contains('routePreviewSummary'));
    expect(rider, contains('Recommended · current traffic'));
    expect(rider, contains("Key('routeOption-\${entry.$2.id}')"));
  });

  test('Google Maps provider types stay behind the shared map boundary', () {
    final rider = File(
      'lib/features/rider_request/presentation/rider_request_screen.dart',
    ).readAsStringSync();
    final driver = File(
      'lib/features/driver_workspace/presentation/driver_workspace_screen.dart',
    ).readAsStringSync();
    final sharedMap = File('lib/core/maps/ride_map.dart').readAsStringSync();

    for (final feature in [rider, driver]) {
      expect(feature, isNot(contains('google_maps_flutter')));
      expect(feature, isNot(contains('flutter_map')));
      expect(feature, isNot(contains('latlong2')));
    }
    expect(sharedMap, contains('google_maps_flutter'));
  });
}
