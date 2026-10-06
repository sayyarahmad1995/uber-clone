import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/rider_request/data/ride_service_repository.dart';

void main() {
  test(
    'Rider catalog contract preserves future service codes and tokens',
    () async {
      final adapter = _RideServicesAdapter();
      final repository = ApiRideServiceRepository(
        Dio(BaseOptions(baseUrl: 'http://application.test'))
          ..httpClientAdapter = adapter,
        _SessionStore(),
      );

      final services = await repository.list();
      expect(adapter.authorization, 'Bearer catalog-token');
      expect(adapter.path, '/v1/ride-services');
      expect(services.map((value) => value.code), [
        'economy',
        'third_car',
        'comfort',
      ]);
      expect(services[1].presentationToken, 'future-icon');
      expect(services[1].displayOrder, 15);
    },
  );
}

class _SessionStore implements SessionStore {
  @override
  Future<String?> readValidToken() async => 'catalog-token';

  @override
  Future<void> save(String token, DateTime expiresAt) async {}

  @override
  Future<void> clear() async {}
}

class _RideServicesAdapter implements HttpClientAdapter {
  String? path;
  String? authorization;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    path = options.path;
    authorization = options.headers['Authorization'] as String?;
    return ResponseBody.fromString(
      jsonEncode({
        'services': [
          {
            'code': 'economy',
            'display_name': 'Economy',
            'description': 'Standard',
            'display_order': 10,
            'presentation_token': 'car',
          },
          {
            'code': 'third_car',
            'display_name': 'Third category',
            'description': 'Test-only service',
            'display_order': 15,
            'presentation_token': 'future-icon',
          },
          {
            'code': 'comfort',
            'display_name': 'Comfort',
            'description': 'Another service',
            'display_order': 20,
            'presentation_token': 'car-front',
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
