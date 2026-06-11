import 'package:dio/dio.dart';

import '../../storage/secure_storage.dart';

/// Добавляет Bearer-токен к каждому запросу.
/// При 401 пробует обновить токен через /auth/refresh и повторяет запрос.
class AuthInterceptor extends Interceptor {
  AuthInterceptor(this._storage, this._dio);

  final SecureStorage _storage;
  final Dio _dio;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _storage.getAccessToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode != 401) {
      handler.next(err);
      return;
    }

    // Пробуем обновить токен
    final refreshToken = await _storage.getRefreshToken();
    if (refreshToken == null) {
      handler.next(err);
      return;
    }

    try {
      final resp = await _dio.post<Map<String, dynamic>>(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
        options: Options(extra: {'skipAuthInterceptor': true}),
      );
      final data = resp.data?['data'] as Map<String, dynamic>?;
      if (data == null) {
        handler.next(err);
        return;
      }

      await _storage.saveTokens(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
      );

      // Повторяем оригинальный запрос с новым токеном
      final opts = err.requestOptions
        ..headers['Authorization'] = 'Bearer ${data['access_token']}';
      final retryResp = await _dio.fetch<dynamic>(opts);
      handler.resolve(retryResp);
    } catch (_) {
      await _storage.clearTokens();
      handler.next(err);
    }
  }
}
