import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';
import 'package:uber_clone/features/rider_request/application/rider_route_preview_controller.dart';
import 'package:uber_clone/features/rider_request/data/route_preview_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';
import 'package:uber_clone/features/rider_request/domain/route_preview.dart';

import 'test_doubles.dart';

void main() {
  test('route preview API returns normalized route options', () async {
    final adapter = _RoutePreviewAdapter();
    final repository = ApiRoutePreviewRepository(
      Dio(BaseOptions(baseUrl: 'http://application.test'))
        ..httpClientAdapter = adapter,
      _SessionStore(),
    );

    final preview = await repository.preview(
      pickup: const GeoPoint(latitude: 38.5, longitude: -120.2),
      destination: const GeoPoint(latitude: 43.252, longitude: -126.453),
      serviceCode: 'economy',
    );

    expect(adapter.path, '/v1/ride-previews');
    expect(adapter.authorization, 'Bearer route-token');
    expect(adapter.serviceCode, 'economy');
    expect(preview.routes, hasLength(2));
    expect(preview.recommended.id, 'route-0');
    expect(preview.recommended.distanceMeters, 788906);
    expect(preview.recommended.durationSeconds, 3600);
    expect(preview.recommended.points, hasLength(3));
    expect(preview.recommended.points.first.latitude, closeTo(38.5, 0.00001));
    expect(preview.recommended.points.first.longitude, closeTo(-120.2, 0.00001));
    expect(preview.routes[1].recommended, isFalse);
  });

  test('encoded route polyline rejects truncated provider data', () {
    expect(
      () => decodeEncodedPolyline('_p~iF'),
      throwsA(isA<FormatException>()),
    );
  });

  test('route preview rejects missing or duplicate recommendation', () {
    expect(
      () => RoutePreview(
        routes: [
          _route(id: 'route-0', recommended: false),
          _route(id: 'route-1', recommended: false),
        ],
      ),
      throwsA(isA<FormatException>()),
    );
    expect(
      () => RoutePreview(
        routes: [
          _route(id: 'route-0', recommended: true),
          _route(id: 'route-1', recommended: true),
        ],
      ),
      throwsA(isA<FormatException>()),
    );
  });

  test('late route response cannot replace a newer preview', () async {
    final rider = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    addTearDown(rider.dispose);

    final repository = _FakeRoutePreviewRepository()
      ..firstPreview = Completer<RoutePreview>();
    final controller = RiderRoutePreviewController(repository, rider);
    addTearDown(controller.dispose);

    rider.setPickup(const GeoPoint(latitude: 24.86, longitude: 67.01));
    rider.setDestination(const GeoPoint(latitude: 24.90, longitude: 67.05));
    await Future<void>.delayed(Duration.zero);

    rider.setDestination(const GeoPoint(latitude: 24.95, longitude: 67.10));
    await Future<void>.delayed(Duration.zero);

    expect(controller.state.selectedRoute?.distanceMeters, 2000);

    repository.firstPreview!.complete(
      RoutePreview(
        routes: [
          _route(
            id: 'old-route',
            recommended: true,
            distanceMeters: 1000,
            durationSeconds: 120,
          ),
        ],
      ),
    );
    await Future<void>.delayed(Duration.zero);

    expect(controller.state.selectedRoute?.distanceMeters, 2000);
  });

  test('recommended route is selected by default and Rider can switch', () async {
    final rider = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    addTearDown(rider.dispose);

    final repository = _FakeRoutePreviewRepository();
    final controller = RiderRoutePreviewController(repository, rider);
    addTearDown(controller.dispose);

    rider.setPickup(const GeoPoint(latitude: 24.86, longitude: 67.01));
    rider.setDestination(const GeoPoint(latitude: 24.90, longitude: 67.05));
    await Future<void>.delayed(Duration.zero);

    expect(controller.state.selectedRoute?.id, 'route-0');
    controller.selectRoute('route-1');
    expect(controller.state.selectedRoute?.id, 'route-1');
    expect(controller.state.selectedRoute?.distanceMeters, 1800);
  });

  test('service change invalidates and reloads the route preview', () async {
    final rider = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    addTearDown(rider.dispose);

    final repository = _FakeRoutePreviewRepository();
    final controller = RiderRoutePreviewController(repository, rider);
    addTearDown(controller.dispose);

    rider.setPickup(const GeoPoint(latitude: 24.86, longitude: 67.01));
    rider.setDestination(const GeoPoint(latitude: 24.90, longitude: 67.05));
    await Future<void>.delayed(Duration.zero);

    rider.selectService('comfort');
    await Future<void>.delayed(Duration.zero);

    expect(repository.serviceCodes, ['economy', 'comfort']);
    expect(controller.state.preview, isNotNull);
    expect(controller.state.selectedRoute?.recommended, isTrue);
  });
}

RouteOption _route({
  required String id,
  required bool recommended,
  int distanceMeters = 2000,
  int durationSeconds = 300,
  String encodedPolyline = '_p~iF~ps|U_ulLnnqC',
}) {
  return RouteOption(
    id: id,
    recommended: recommended,
    distanceMeters: distanceMeters,
    durationSeconds: durationSeconds,
    encodedPolyline: encodedPolyline,
  );
}

class _FakeRoutePreviewRepository implements RoutePreviewRepository {
  Completer<RoutePreview>? firstPreview;
  final serviceCodes = <String>[];
  var calls = 0;

  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    required GeoPoint destination,
    required String serviceCode,
  }) {
    calls++;
    serviceCodes.add(serviceCode);
    if (calls == 1 && firstPreview != null) {
      return firstPreview!.future;
    }
    return Future.value(
      RoutePreview(
        routes: [
          _route(id: 'route-0', recommended: true),
          _route(
            id: 'route-1',
            recommended: false,
            distanceMeters: 1800,
            durationSeconds: 340,
            encodedPolyline: '_p~iF~ps|U????',
          ),
        ],
      ),
    );
  }

  @override
  void cancel() {}
}

class _SessionStore implements SessionStore {
  @override
  Future<String?> readValidToken() async => 'route-token';

  @override
  Future<void> save(String token, DateTime expiresAt) async {}

  @override
  Future<void> clear() async {}
}

class _RoutePreviewAdapter implements HttpClientAdapter {
  String? path;
  String? authorization;
  String? serviceCode;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    path = options.path;
    authorization = options.headers['Authorization'] as String?;
    final data = options.data as Map<String, dynamic>;
    serviceCode = data['service_code'] as String?;

    return ResponseBody.fromString(
      jsonEncode({
        'routes': [
          {
            'id': 'route-0',
            'recommended': true,
            'distance_meters': 788906,
            'duration_seconds': 3600,
            'encoded_polyline': '_p~iF~ps|U????',
          },
          {
            'id': 'route-1',
            'recommended': false,
            'distance_meters': 800000,
            'duration_seconds': 3900,
            'encoded_polyline': '_p~iF~ps|U_ulLnnqC',
          },
        ],
      }),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
