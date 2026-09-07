import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:uber_clone/core/network/api_exception.dart';

void main() {
  test('backend error text is preserved when message is absent', () {
    final request = RequestOptions(path: '/v1/driver/onboarding');
    final exception = ApiException.fromDio(
      DioException(
        requestOptions: request,
        response: Response<Map<String, dynamic>>(
          requestOptions: request,
          statusCode: 409,
          data: {
            'error': 'a Driver onboarding application is already under review',
          },
        ),
      ),
    );

    expect(exception.statusCode, 409);
    expect(
      exception.message,
      'a Driver onboarding application is already under review',
    );
  });
}
