import 'dart:developer' as dev;

import 'package:dio/dio.dart';

/// Интерсептор для логгирования запросов и ответов в dev режиме.
class LoggingInterceptor extends Interceptor {
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    dev.log(
      '→ ${options.method} ${options.uri}',
      name: 'API',
    );
    super.onRequest(options, handler);
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    dev.log(
      '← ${response.statusCode} ${response.requestOptions.uri}',
      name: 'API',
    );
    super.onResponse(response, handler);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    dev.log(
      '✗ ${err.response?.statusCode ?? '???'} ${err.requestOptions.uri}: ${err.message}',
      name: 'API',
      error: err,
    );
    super.onError(err, handler);
  }
}
