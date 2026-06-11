import 'package:dio/dio.dart';

import '../config/app_config.dart';
import 'interceptors/logging_interceptor.dart';

/// Создаёт настроенный экземпляр Dio.
/// Будет использоваться через Riverpod провайдер.
Dio createDio() {
  final dio = Dio(
    BaseOptions(
      baseUrl: AppConfig.apiUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      sendTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      validateStatus: (status) => status != null && status >= 200 && status < 300,
    ),
  );

  if (AppConfig.isDev) {
    dio.interceptors.add(LoggingInterceptor());
  }

  // На Этапе 1 здесь будет добавлен AuthInterceptor.
  // dio.interceptors.add(AuthInterceptor());

  return dio;
}
