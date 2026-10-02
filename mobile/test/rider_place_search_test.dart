import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/rider_request/application/rider_place_search_controller.dart';
import 'package:uber_clone/features/rider_request/application/rider_request_controller.dart';
import 'package:uber_clone/features/rider_request/data/place_search_repository.dart';
import 'package:uber_clone/features/rider_request/domain/place_search.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

import 'test_doubles.dart';

void main() {
  test('API place search contract uses auth and app-owned endpoints', () async {
    final adapter = _PlacesAdapter();
    final repository = ApiPlaceSearchRepository(
      Dio(BaseOptions(baseUrl: 'http://application.test'))
        ..httpClientAdapter = adapter,
      _SessionStore(),
    );

    final suggestions = await repository.autocomplete(
      input: 'Clifton',
      sessionToken: 'session-1',
      bias: const GeoPoint(latitude: 24.86, longitude: 67.01),
    );
    final details = await repository.details(
      placeId: suggestions.single.placeId,
      sessionToken: 'session-1',
    );
    final reverse = await repository.reverseGeocode(
      const GeoPoint(latitude: 24.87, longitude: 67.02),
    );

    expect(adapter.authorizations, everyElement('Bearer places-token'));
    expect(adapter.paths, [
      '/v1/places/autocomplete',
      '/v1/places/p1',
      '/v1/places/reverse-geocode',
    ]);
    expect(adapter.autocompleteSessionToken, 'session-1');
    expect(adapter.detailsSessionToken, 'session-1');
    expect(details.label, 'Clifton, Karachi');
    expect(reverse!.label, 'Pinned Road, Karachi');
  });

  test(
    'pickup and destination use independent autocomplete sessions',
    () async {
      final rider = RiderRequestController(
        FakeRideRequestRepository(),
        const FakeDeviceLocation(),
      );
      addTearDown(rider.dispose);
      final places = _FakePlaceSearchRepository();
      final controller = RiderPlaceSearchController(
        places,
        rider,
        const FakeDeviceLocation(),
      );
      addTearDown(controller.dispose);

      controller.search(RiderPlaceField.pickup, 'Clifton');
      controller.search(RiderPlaceField.destination, 'Airport');
      await Future<void>.delayed(const Duration(milliseconds: 350));

      expect(places.autocompleteCalls, hasLength(2));
      final pickupToken = places.autocompleteCalls
          .firstWhere((call) => call.input == 'Clifton')
          .sessionToken;
      final destinationToken = places.autocompleteCalls
          .firstWhere((call) => call.input == 'Airport')
          .sessionToken;
      expect(pickupToken, isNot(destinationToken));

      await controller.select(
        RiderPlaceField.pickup,
        const PlaceSuggestion(placeId: 'pickup-1', label: 'Clifton'),
      );
      expect(places.detailsTokens.single, pickupToken);

      controller.search(RiderPlaceField.pickup, 'Sea View');
      await Future<void>.delayed(const Duration(milliseconds: 350));
      final secondPickupToken = places.autocompleteCalls
          .firstWhere((call) => call.input == 'Sea View')
          .sessionToken;
      expect(secondPickupToken, isNot(pickupToken));
    },
  );

  test('late autocomplete result cannot replace a newer query', () async {
    final rider = RiderRequestController(
      FakeRideRequestRepository(),
      const FakeDeviceLocation(),
    );
    addTearDown(rider.dispose);
    final places = _FakePlaceSearchRepository()
      ..firstAutocomplete = Completer<List<PlaceSuggestion>>();
    final controller = RiderPlaceSearchController(
      places,
      rider,
      const FakeDeviceLocation(),
    );
    addTearDown(controller.dispose);

    controller.search(RiderPlaceField.destination, 'First place');
    await Future<void>.delayed(const Duration(milliseconds: 350));
    controller.search(RiderPlaceField.destination, 'Second place');
    await Future<void>.delayed(const Duration(milliseconds: 350));

    expect(
      controller.state.destination.suggestions.single.label,
      'Second place result',
    );

    places.firstAutocomplete!.complete(const [
      PlaceSuggestion(placeId: 'old', label: 'First place result'),
    ]);
    await Future<void>.delayed(Duration.zero);

    expect(
      controller.state.destination.suggestions.single.label,
      'Second place result',
    );
  });

  test(
    'reverse geocode always keeps the exact Rider pin',
    () async {
      final rider = RiderRequestController(
        FakeRideRequestRepository(),
        const FakeDeviceLocation(),
      );
      addTearDown(rider.dispose);
      final places = _FakePlaceSearchRepository();
      final controller = RiderPlaceSearchController(
        places,
        rider,
        const FakeDeviceLocation(),
      );
      addTearDown(controller.dispose);

      const pin = GeoPoint(latitude: 24.86123, longitude: 67.01987);
      final resolved = await controller.reconcilePin(
        RiderPlaceField.pickup,
        pin,
      );

      expect(resolved!.label, 'Pinned Road, Karachi');
      expect(resolved.point, pin);
      expect(rider.state.pickup, pin);
      expect(controller.state.pickup.label, 'Pinned Road, Karachi');
    },
  );
}

class _AutocompleteCall {
  const _AutocompleteCall(this.input, this.sessionToken);

  final String input;
  final String sessionToken;
}

class _FakePlaceSearchRepository implements PlaceSearchRepository {
  final autocompleteCalls = <_AutocompleteCall>[];
  final detailsTokens = <String>[];
  Completer<List<PlaceSuggestion>>? firstAutocomplete;
  PlaceSelection reverseSelection = const PlaceSelection(
    placeId: 'pin-1',
    label: 'Pinned Road, Karachi',
    point: GeoPoint(latitude: 24.86, longitude: 67.02),
  );

  @override
  Future<List<PlaceSuggestion>> autocomplete({
    required String input,
    required String sessionToken,
    GeoPoint? bias,
  }) {
    autocompleteCalls.add(_AutocompleteCall(input, sessionToken));
    if (input == 'First place' && firstAutocomplete != null) {
      return firstAutocomplete!.future;
    }
    return Future.value([
      PlaceSuggestion(
        placeId: input.toLowerCase().replaceAll(' ', '-'),
        label: '$input result',
      ),
    ]);
  }

  @override
  Future<PlaceSelection> details({
    required String placeId,
    required String sessionToken,
  }) async {
    detailsTokens.add(sessionToken);
    return const PlaceSelection(
      placeId: 'pickup-1',
      label: 'Clifton, Karachi',
      point: GeoPoint(latitude: 24.82, longitude: 67.03),
    );
  }

  @override
  Future<PlaceSelection?> reverseGeocode(GeoPoint point) async =>
      reverseSelection;

  @override
  void cancelSession(String sessionToken) {}
}

class _SessionStore implements SessionStore {
  @override
  Future<String?> readValidToken() async => 'places-token';

  @override
  Future<void> save(String token, DateTime expiresAt) async {}

  @override
  Future<void> clear() async {}
}

class _PlacesAdapter implements HttpClientAdapter {
  final paths = <String>[];
  final authorizations = <String?>[];
  String? autocompleteSessionToken;
  String? detailsSessionToken;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    paths.add(options.path);
    authorizations.add(options.headers['Authorization'] as String?);

    if (options.path == '/v1/places/autocomplete') {
      final data = options.data as Map<String, dynamic>;
      autocompleteSessionToken = data['session_token'] as String?;
      return _json({
        'suggestions': [
          {'place_id': 'p1', 'label': 'Clifton, Karachi'},
        ],
      });
    }
    if (options.path == '/v1/places/p1') {
      detailsSessionToken = options.queryParameters['session_token'] as String?;
      return _json({
        'place_id': 'p1',
        'label': 'Clifton, Karachi',
        'latitude': 24.82,
        'longitude': 67.03,
      });
    }
    if (options.path == '/v1/places/reverse-geocode') {
      return _json({
        'place_id': 'pin-1',
        'label': 'Pinned Road, Karachi',
        'latitude': 24.87,
        'longitude': 67.02,
      });
    }
    return ResponseBody.fromString(
      jsonEncode({'error': 'not found'}),
      404,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  ResponseBody _json(Map<String, dynamic> body) => ResponseBody.fromString(
    jsonEncode(body),
    200,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );

  @override
  void close({bool force = false}) {}
}
