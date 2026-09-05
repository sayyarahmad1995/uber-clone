import 'package:dio/dio.dart';

import '../../../core/models/account.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/session/session_store.dart';

abstract interface class AuthRepository {
  Future<Account?> restore();
  Future<Account> login(String identifier, String password);
  Future<String> register(String identifier, String password);
  Future<void> completeVerification(String verificationId, String code);
  Future<void> logout();
  Future<Account> enableDriver();
}

class ApiAuthRepository implements AuthRepository {
  ApiAuthRepository(this._dio, this._sessions);
  final Dio _dio;
  final SessionStore _sessions;

  @override
  Future<Account> enableDriver() async {
    final token = await _sessions.readValidToken();
    if (token == null) {
      throw const ApiException(
        'authentication_required',
        'Please sign in again.',
        statusCode: 401,
      );
    }
    try {
      final response = await _dio.put<Map<String, dynamic>>(
        '/v1/me/capabilities/driver',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );
      return Account.fromJson(response.data!);
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<Account?> restore() async {
    final token = await _sessions.readValidToken();
    if (token == null) return null;
    try {
      return await _loadAccount(token);
    } on ApiException catch (error) {
      if (error.isUnauthorized) {
        await _sessions.clear();
        return null;
      }
      rethrow;
    }
  }

  @override
  Future<Account> login(String identifier, String password) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/auth/login',
        data: {'identifier': identifier.trim(), 'password': password},
      );
      final session = AuthSession.fromJson(response.data!);
      await _sessions.save(
        session.accessToken,
        DateTime.now().add(Duration(seconds: session.expiresIn)),
      );
      try {
        return await _loadAccount(session.accessToken);
      } catch (_) {
        await _sessions.clear();
        rethrow;
      }
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<String> register(String identifier, String password) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/auth/register',
        data: {'identifier': identifier.trim(), 'password': password},
      );
      return VerificationChallenge.fromJson(response.data!).verificationId;
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<void> completeVerification(String verificationId, String code) async {
    try {
      await _dio.post<void>(
        '/v1/auth/verify/complete',
        data: {'verification_id': verificationId, 'code': code.trim()},
      );
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  Future<Account> _loadAccount(String token) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/v1/me',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );
      return Account.fromJson(response.data!);
    } on DioException catch (error) {
      throw ApiException.fromDio(error);
    }
  }

  @override
  Future<void> logout() async {
    final token = await _sessions.readValidToken();
    try {
      if (token != null) {
        await _dio.post<void>(
          '/v1/auth/logout',
          options: Options(headers: {'Authorization': 'Bearer $token'}),
        );
      }
    } on DioException catch (error) {
      if (error.response?.statusCode != 401) throw ApiException.fromDio(error);
    } finally {
      await _sessions.clear();
    }
  }
}
