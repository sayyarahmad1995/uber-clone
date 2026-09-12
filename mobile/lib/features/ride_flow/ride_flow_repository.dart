import 'package:dio/dio.dart';
import '../../core/network/api_exception.dart';
import '../../core/session/session_store.dart';

typedef Json = Map<String, dynamic>;

abstract interface class RideFlowRepository {
  Future<Json> get(String path);
  Future<void> act(String path, {Json? data, bool put = false});
}

class ApiRideFlowRepository implements RideFlowRepository {
  ApiRideFlowRepository(this.dio, this.sessions);
  final Dio dio;
  final SessionStore sessions;
  Future<Json> _request(String path, String method, Json? data) async {
    final token = await sessions.readValidToken();
    if (token == null) {
      throw const ApiException('authentication_required', 'Please sign in again.', statusCode: 401);
    }
    try {
      final response = await dio.request<Json>(path, data: data,
        options: Options(method: method, headers: {'Authorization': 'Bearer $token'}));
      return response.data ?? <String, dynamic>{};
    } on DioException catch (error) { throw ApiException.fromDio(error); }
  }
  @override
  Future<Json> get(String path) => _request(path, 'GET', null);
  @override
  Future<void> act(String path, {Json? data, bool put = false}) async {
    await _request(path, put ? 'PUT' : 'POST', data);
  }
}
