import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/features/ride_flow/ride_flow_repository.dart';

import 'ride_request_repository_test.dart' show StaticSessionStore;

const assignedOperationContext = <String, dynamic>{
  'driver_name': 'Ayesha Khan',
  'vehicle_id': '11111111-1111-1111-1111-111111111111',
  'make': 'Toyota',
  'model': 'Corolla',
  'model_year': 2024,
  'color': 'White',
  'license_plate': 'XYZ 987',
  'service_code': 'comfort',
  'service_name': 'Comfort',
  'fare': {'amount_minor': 125000, 'currency': 'PKR'},
};

void main() {
  late RecordingRideFlowAdapter adapter;
  late ApiRideFlowRepository repository;

  setUp(() {
    adapter = RecordingRideFlowAdapter();
    final dio = Dio(BaseOptions(baseUrl: 'http://application.test'))
      ..httpClientAdapter = adapter;
    repository = ApiRideFlowRepository(dio, StaticSessionStore());
  });

  test('list marketplace requests uses the marketplace route', () async {
    await repository.listMarketplaceRequests();

    expectRequest(adapter, 'GET', '/v1/driver/marketplace/ride-requests');
  });

  test('Driver polling parses the complete assigned trip contract', () async {
    final trip = await repository.getCurrentDriverTrip();

    expectRequest(adapter, 'GET', '/v1/driver/trip');
    expect(trip?.operationContext?.licensePlate, 'XYZ 987');
    expect(trip?.operationContext?.serviceCode, 'comfort');
  });

  test('list driver trips uses the driver trips route', () async {
    await repository.listDriverTrips();

    expectRequest(adapter, 'GET', '/v1/driver/trips');
  });

  test('list rider rides uses the rider rides route', () async {
    await repository.listRiderRides();

    expectRequest(adapter, 'GET', '/v1/ride-requests');
  });

  test('Rider polling parses the complete assigned trip contract', () async {
    final ride = await repository.getRiderRide('ride-1');

    expectRequest(adapter, 'GET', '/v1/ride-requests/ride-1');
    expect(ride.trip?.operationContext?.licensePlate, 'XYZ 987');
    expect(ride.trip?.operationContext?.serviceCode, 'comfort');
  });

  test('list rider offers uses the offers route', () async {
    await repository.listRiderOffers('ride-1');

    expectRequest(adapter, 'GET', '/v1/ride-requests/ride-1/offers');
  });

  test('get driver location uses the driver location route', () async {
    await repository.getDriverLocation('ride-1');

    expectRequest(adapter, 'GET', '/v1/ride-requests/ride-1/driver-location');
  });

  test('submit offer uses the offer route and amount payload', () async {
    await repository.submitOffer('ride-1', 11500);

    expectRequest(adapter, 'PUT', '/v1/driver/ride-requests/ride-1/offer', {
      'amount_minor': 11500,
    });
  });

  test('accept proposed fare uses existing backend accept route', () async {
    await repository.acceptProposedFare('ride-1');

    expect(adapter.request!.method, 'POST');
    expect(adapter.request!.path, '/v1/driver/ride-requests/ride-1/accept');
    expect(adapter.request!.data, isNull);
  });

  test('Driver decline uses the opportunity skip route', () async {
    await repository.declineRideRequest('ride-1');

    expectRequest(adapter, 'POST', '/v1/driver/ride-requests/ride-1/skip');
  });

  test('Rider decline uses the specific offer rejection route', () async {
    await repository.declineOffer('ride-1', 'driver-1');

    expectRequest(
      adapter,
      'POST',
      '/v1/ride-requests/ride-1/offers/driver-1/reject',
    );
  });

  test(
    'select offer uses the offer acceptance route and UTC payload',
    () async {
      final updatedAt = DateTime(2026, 9, 17, 14, 30);

      await repository.selectOffer('ride-1', 'driver-1', updatedAt);

      expectRequest(
        adapter,
        'POST',
        '/v1/ride-requests/ride-1/offers/driver-1/accept',
        {'updated_at': updatedAt.toUtc().toIso8601String()},
      );
    },
  );

  test('start trip uses the start route', () async {
    await repository.startTrip('ride-1');

    expectRequest(adapter, 'POST', '/v1/driver/ride-requests/ride-1/start');
  });

  test('complete trip uses the complete route', () async {
    await repository.completeTrip('ride-1');

    expectRequest(adapter, 'POST', '/v1/driver/ride-requests/ride-1/complete');
  });

  test('confirm cash collected uses the cash-collected route', () async {
    await repository.confirmCashCollected('ride-1');

    expectRequest(
      adapter,
      'POST',
      '/v1/driver/ride-requests/ride-1/cash-collected',
    );
  });

  test('cancel driver trip uses the cancel route', () async {
    await repository.cancelDriverTrip('ride-1');

    expectRequest(adapter, 'POST', '/v1/driver/ride-requests/ride-1/cancel');
  });
}

void expectRequest(
  RecordingRideFlowAdapter adapter,
  String method,
  String path, [
  Map<String, dynamic>? data,
]) {
  expect(adapter.request!.method, method);
  expect(adapter.request!.path, path);
  expect(adapter.request!.data, data);
}

class RecordingRideFlowAdapter implements HttpClientAdapter {
  RequestOptions? request;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    request = options;
    final body = switch (options.path) {
      '/v1/driver/marketplace/ride-requests' => {'ride_requests': []},
      '/v1/driver/trip' => {
        'operation_context': assignedOperationContext,
        'ride_request_id': 'ride-1',
        'rider_user_id': 'rider-1',
        'driver_user_id': 'driver-1',
        'status': 'assigned',
        'assigned_at': '2026-09-17T10:00:00Z',
        'started_at': null,
        'completed_at': null,
        'cancelled_at': null,
        'settlement': {
          'status': 'unsettled',
          'method': null,
          'cash_collected_at': null,
        },
      },
      '/v1/driver/trips' => {'trips': []},
      '/v1/ride-requests' => {'ride_requests': []},
      '/v1/ride-requests/ride-1' => {
        'service_code': 'standard',
        'id': 'ride-1',
        'pickup': {'latitude': 24.86, 'longitude': 67.00},
        'destination': {'latitude': 24.90, 'longitude': 67.08},
        'proposed_fare': {'amount_minor': 11500, 'currency': 'PKR'},
        'status': 'accepted',
        'created_at': '2026-09-17T10:00:00Z',
        'expires_at': '2026-09-17T10:15:00Z',
        'trip': {
          'operation_context': assignedOperationContext,
          'driver_user_id': 'driver-1',
          'status': 'assigned',
          'assigned_at': '2026-09-17T10:01:00Z',
          'started_at': null,
          'completed_at': null,
          'cancelled_at': null,
          'settlement': {
            'status': 'unsettled',
            'method': null,
            'cash_collected_at': null,
          },
        },
      },
      '/v1/ride-requests/ride-1/offers' => {'offers': []},
      '/v1/ride-requests/ride-1/driver-location' => {
        'ride_request_id': 'ride-1',
        'latitude': 24.86,
        'longitude': 67.00,
        'updated_at': '2026-09-17T10:00:00Z',
      },
      _ => <String, dynamic>{},
    };
    return ResponseBody.fromString(
      jsonEncode(body),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
