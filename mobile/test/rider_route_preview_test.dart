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
  test('route preview API returns one recommended route', () async {
    final adapter = _RoutePreviewAdapter();
    final repository = ApiRoutePreviewRepository(
      Dio(BaseOptions(baseUrl: 'http://application.test'))
        ..httpClientAdapter = adapter,
      _SessionStore(),
    );

    final preview = await repository.preview(
      pickup: const GeoPoint(latitude: 38.5, longitude: -120.2),
      destination: const GeoPoint(latitude: 43.252, longitude: -126.453),
      destinationPlaceId: 'market-place-id',
      serviceCode: 'economy',
    );

    expect(adapter.path, '/v1/ride-previews');
    expect(adapter.authorization, 'Bearer route-token');
    expect(adapter.serviceCode, 'economy');
    expect(adapter.pickupPlaceId, isNull);
    expect(adapter.destinationPlaceId, 'market-place-id');
    expect(preview.distanceMeters, 788906);
    expect(preview.durationSeconds, 3600);
    expect(preview.suggestedFare.amountMinor, 34500);
    expect(preview.suggestedFare.currency, 'PKR');
    expect(preview.pricingPolicyVersion, 4);
    expect(preview.points, hasLength(3));
    expect(preview.points.first.latitude, closeTo(38.5, 0.00001));
    expect(preview.points.first.longitude, closeTo(-120.2, 0.00001));
  });

  test('encoded route polyline rejects truncated provider data', () {
    expect(
      () => decodeEncodedPolyline('_p~iF'),
      throwsA(isA<FormatException>()),
    );
  });

  test('route preview rejects missing route object', () {
    expect(
      () => RoutePreview.fromJson(const {}),
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

    rider.setPickup(
      const GeoPoint(latitude: 24.86, longitude: 67.01),
      placeId: 'pickup-place',
    );
    rider.setDestination(
      const GeoPoint(latitude: 24.90, longitude: 67.05),
      placeId: 'destination-place',
    );
    await Future<void>.delayed(Duration.zero);

    rider.setDestination(const GeoPoint(latitude: 24.95, longitude: 67.10));
    await Future<void>.delayed(Duration.zero);

    expect(controller.state.preview?.distanceMeters, 2000);

    repository.firstPreview!.complete(
      _route(distanceMeters: 1000, durationSeconds: 120),
    );
    await Future<void>.delayed(Duration.zero);

    expect(controller.state.preview?.distanceMeters, 2000);
  });

  test('Place ID change invalidates preview even when coordinates stay the same', () async {
    final rider = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    addTearDown(rider.dispose);

    final repository = _FakeRoutePreviewRepository();
    final controller = RiderRoutePreviewController(repository, rider);
    addTearDown(controller.dispose);

    const point = GeoPoint(latitude: 24.86, longitude: 67.01);
    rider.setPickup(point, placeId: 'pickup-a');
    rider.setDestination(
      const GeoPoint(latitude: 24.90, longitude: 67.05),
      placeId: 'destination-a',
    );
    await Future<void>.delayed(Duration.zero);

    rider.setPickup(point, placeId: 'pickup-b');
    await Future<void>.delayed(Duration.zero);

    expect(repository.pickupPlaceIds, ['pickup-a', 'pickup-b']);
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
    expect(controller.state.preview?.distanceMeters, 2000);
  });
}

RoutePreview _route({
  int distanceMeters = 2000,
  int durationSeconds = 300,
  String encodedPolyline = '_p~iF~ps|U_ulLnnqC',
}) {
  return RoutePreview(
    distanceMeters: distanceMeters,
    durationSeconds: durationSeconds,
    encodedPolyline: encodedPolyline,
    suggestedFare: const Money(amountMinor: 25000, currency: 'PKR'),
    pricingPolicyVersion: 2,
  );
}

class _FakeRoutePreviewRepository implements RoutePreviewRepository {
  Completer<RoutePreview>? firstPreview;
  final serviceCodes = <String>[];
  final pickupPlaceIds = <String?>[];
  final destinationPlaceIds = <String?>[];
  var calls = 0;

  @override
  Future<RoutePreview> preview({
    required GeoPoint pickup,
    String? pickupPlaceId,
    required GeoPoint destination,
    String? destinationPlaceId,
    required String serviceCode,
  }) {
    calls++;
    serviceCodes.add(serviceCode);
    pickupPlaceIds.add(pickupPlaceId);
    destinationPlaceIds.add(destinationPlaceId);
    if (calls == 1 && firstPreview != null) {
      return firstPreview!.future;
    }
    return Future.value(_route());
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
  String? pickupPlaceId;
  String? destinationPlaceId;

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
    pickupPlaceId =
        (data['pickup'] as Map<String, dynamic>)['place_id'] as String?;
    destinationPlaceId =
        (data['destination'] as Map<String, dynamic>)['place_id'] as String?;

    return ResponseBody.fromString(
      jsonEncode({
        'route': {
          'distance_meters': 788906,
          'duration_seconds': 3600,
          'encoded_polyline': '_p~iF~ps|U????',
        },
        'suggested_fare': {
          'amount_minor': 34500,
          'currency': 'PKR',
        },
        'pricing_policy_version': 4,
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
