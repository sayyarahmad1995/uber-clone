import 'package:dio/dio.dart';

class ApiException implements Exception {
  const ApiException(this.code, this.message, {this.statusCode});
  final String code;
  final String message;
  final int? statusCode;
  bool get isUnauthorized => statusCode == 401;

  factory ApiException.fromDio(DioException error) {
    final data = error.response?.data;
    final body = data is Map<String, dynamic> ? data : null;
    return ApiException(
      body?['error'] as String? ?? 'request_failed',
      body?['message'] as String? ?? 'Unable to complete the request.',
      statusCode: error.response?.statusCode,
    );
  }

  @override
  String toString() => message;
}
