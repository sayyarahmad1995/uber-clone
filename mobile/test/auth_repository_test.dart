import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/session/session_store.dart';
import 'package:uber_clone/features/authentication/data/auth_repository.dart';

void main() {
  test('login stores the session and loads the Rider account', () async {
    final adapter = ContractAdapter();
    final sessions = MemorySessionStore();
    final repository = ApiAuthRepository(
      Dio(BaseOptions(baseUrl: 'http://application.test'))
        ..httpClientAdapter = adapter,
      sessions,
    );

    final account = await repository.login(' rider@example.com ', 'secret');

    expect(account.id, 'user-1');
    expect(account.capabilities.single.name, 'rider');
    expect(sessions.token, 'session-token');
    expect(adapter.loginBody, {
      'identifier': 'rider@example.com',
      'password': 'secret',
    });
    expect(adapter.meAuthorization, 'Bearer session-token');
  });
}

class MemorySessionStore implements SessionStore {
  String? token;
  DateTime? expiresAt;
  @override
  Future<void> clear() async {
    token = null;
    expiresAt = null;
  }

  @override
  Future<String?> readValidToken() async => token;
  @override
  Future<void> save(String value, DateTime expiry) async {
    token = value;
    expiresAt = expiry;
  }
}

class ContractAdapter implements HttpClientAdapter {
  Map<String, dynamic>? loginBody;
  String? meAuthorization;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path == '/v1/auth/login') {
      loginBody = Map<String, dynamic>.from(options.data as Map);
      return jsonResponse(200, {
        'access_token': 'session-token',
        'expires_in': 3600,
      });
    }
    if (options.path == '/v1/me') {
      meAuthorization = options.headers['Authorization'] as String?;
      return jsonResponse(200, {
        'id': 'user-1',
        'capabilities': ['rider'],
      });
    }
    return jsonResponse(404, {'error': 'not_found'});
  }

  ResponseBody jsonResponse(int status, Map<String, dynamic> body) =>
      ResponseBody.fromString(
        jsonEncode(body),
        status,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );

  @override
  void close({bool force = false}) {}
}
