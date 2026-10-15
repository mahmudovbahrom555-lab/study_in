import 'package:dio/dio.dart';

import '../../storage/secure_storage.dart';

/// Добавляет Bearer-токен к каждому запросу.
/// При 401 обновляет токен через /auth/refresh и повторяет запрос.
class AuthInterceptor extends Interceptor {
  AuthInterceptor(
    this._storage,
    this._dio, {
    required this.onSessionExpired,
    Dio? refreshDio,
  }) : _refreshDio = refreshDio ?? Dio(_dio.options.copyWith());

  final SecureStorage _storage;
  final Dio _dio;

  /// Обновление идёт отдельным клиентом без этого перехватчика: иначе 401 на
  /// самом /auth/refresh снова запускает обновление — бесконечный цикл.
  final Dio _refreshDio;

  /// Сервер отверг refresh-токен (истёк, отозван, аккаунт удалён) —
  /// приложение должно вернуть пользователя на вход.
  final void Function() onSessionExpired;

  /// Одно обновление на все запросы, получившие 401 одновременно: сервер
  /// меняет refresh-токен при каждом обновлении, второй параллельный запрос
  /// со старым токеном выкинул бы пользователя из аккаунта.
  Future<String?>? _refreshing;

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

    final newToken = await (_refreshing ??= _refresh().whenComplete(() => _refreshing = null));
    if (newToken == null) {
      handler.next(err);
      return;
    }
    try {
      final opts = err.requestOptions
        ..headers['Authorization'] = 'Bearer $newToken';
      handler.resolve(await _dio.fetch<dynamic>(opts));
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  /// Новый access-токен или null, если обновить не удалось.
  Future<String?> _refresh() async {
    final refreshToken = await _storage.getRefreshToken();
    if (refreshToken == null) return null;
    try {
      final resp = await _refreshDio.post<Map<String, dynamic>>(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      final data = resp.data?['data'] as Map<String, dynamic>?;
      if (data == null) return null;
      final access = data['access_token'] as String;
      await _storage.saveTokens(
        accessToken: access,
        refreshToken: data['refresh_token'] as String,
      );
      return access;
    } on DioException catch (e) {
      // Выходим, только если сервер отверг refresh-токен.
      // Обрыв сети при обновлении — не повод терять вход.
      final status = e.response?.statusCode;
      if (status == 400 || status == 401) {
        await _storage.clearTokens();
        onSessionExpired();
      }
      return null;
    }
  }
}
