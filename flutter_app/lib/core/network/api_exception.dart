import 'package:dio/dio.dart';

/// Исключение API, унифицирующее ошибки сети и сервера.
class ApiException implements Exception {
  ApiException({
    required this.code,
    required this.message,
    this.statusCode,
    this.details,
  });

  /// Создаёт ApiException из DioException.
  factory ApiException.fromDio(DioException err) {
    if (err.type == DioExceptionType.connectionTimeout ||
        err.type == DioExceptionType.receiveTimeout ||
        err.type == DioExceptionType.sendTimeout) {
      return ApiException(
        code: 'TIMEOUT',
        message: 'Превышено время ожидания',
      );
    }

    if (err.type == DioExceptionType.connectionError) {
      return ApiException(
        code: 'NO_CONNECTION',
        message: 'Нет соединения с сервером',
      );
    }

    final response = err.response;
    if (response == null) {
      return ApiException(
        code: 'UNKNOWN',
        message: err.message ?? 'Неизвестная ошибка',
      );
    }

    // Стандартный формат ошибки сервера:
    // { "error": { "code": "...", "message": "...", "details": {...} } }
    final data = response.data;
    if (data is Map<String, dynamic> && data['error'] is Map<String, dynamic>) {
      final error = data['error'] as Map<String, dynamic>;
      return ApiException(
        code: error['code'] as String? ?? 'UNKNOWN',
        message: error['message'] as String? ?? 'Ошибка сервера',
        statusCode: response.statusCode,
        details: error['details'] as Map<String, dynamic>?,
      );
    }

    return ApiException(
      code: 'SERVER_ERROR',
      message: 'Ошибка сервера',
      statusCode: response.statusCode,
    );
  }

  final String code;
  final String message;
  final int? statusCode;
  final Map<String, dynamic>? details;

  @override
  String toString() => 'ApiException($code): $message';
}
