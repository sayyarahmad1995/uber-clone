import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/rider_request/data/ride_request_repository.dart';
import 'package:uber_clone/features/rider_request/domain/ride_request.dart';

void main() {
  test(
    'creates a unified marketplace request with bearer authentication',
    () async {
      final adapter = RideContractAdapter();
      final repository = ApiRideRequestRepository(
        Dio(BaseOptions(baseUrl: 'http://application.test'))
          ..httpClientAdapter = adapter,
        StaticSessionStore(),
      );

      final result = await repository.create(
        pickup: const GeoPoint(latitude: 24.8607, longitude: 67.0011),
        destination: const GeoPoint(latitude: 24.9056, longitude: 67.0822),
        proposedFare: const Money(amountMinor: 70000, currency: 'PKR'),
      );

      expect(result.id, 'ride-1');
      expect(adapter.authorization, 'Bearer session-token');
      expect(adapter.body?['pickup'], {
        'latitude': 24.8607,
        'longitude': 67.0011,
      });
      expect(adapter.body?['proposed_fare'], {
        'amount_minor': 70000,
        'currency': 'PKR',
      });
      expect(adapter.body?.containsKey('booking_mode'), isFalse);
    },
  );
}

class StaticSessionStore implements SessionStore {
  @override
  Future<String?> readValidToken() async => 'session-token';
  @override
  Future<void> clear() async {}
  @override
  Future<void> save(String token, DateTime expiresAt) async {}
}

class RideContractAdapter implements HttpClientAdapter {
  String? authorization;
  Map<String, dynamic>? body;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    authorization = options.headers['Authorization'] as String?;
    body = Map<String, dynamic>.from(options.data as Map);
    return ResponseBody.fromString(
      jsonEncode({
        'id': 'ride-1',
        'rider_user_id': 'user-1',
        'pickup': {'latitude': 24.8607, 'longitude': 67.0011},
        'destination': {'latitude': 24.9056, 'longitude': 67.0822},
        'proposed_fare': {'amount_minor': 70000, 'currency': 'PKR'},
        'status': 'requested',
        'created_at': '2026-09-04T00:00:00Z',
      }),
      201,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
