import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../localization/l10n.dart';

/// Переводит любую ошибку в понятный пользователю текст на языке интерфейса.
///
/// Никогда не возвращает текст исключения: пользователь не должен видеть
/// `DioException [bad response]…` и системные сообщения.
/// Сырая ошибка печатается в консоль в debug-сборке, чтобы разработчик её видел.
String userErrorMessage(Object error) {
  if (kDebugMode) debugPrint('userErrorMessage: $error');
  final l10n = currentL10n;

  if (error is DioException) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return l10n.errTimeout;
      case DioExceptionType.connectionError:
        return l10n.errNoConnection;
      case DioExceptionType.cancel:
        return l10n.errCancelled;
      case DioExceptionType.badCertificate:
        return l10n.errBadCertificate;
      case DioExceptionType.badResponse:
        return _byStatus(
          l10n,
          error.response?.statusCode,
          _serverMessage(error.response?.data),
        );
      case DioExceptionType.unknown:
        return l10n.errGeneric;
    }
  }
  return l10n.errGeneric;
}

String _byStatus(AppLocalizations l10n, int? status, String? serverMessage) {
  // Сервер пока отвечает то по-русски («Роль уже установлена»), то по-английски
  // («forbidden»). Его текст показываем, только если он на языке интерфейса
  // и написан для людей — иначе берём свой перевод по коду ответа.
  final readable = _matchesLanguage(serverMessage, l10n.localeName)
      ? serverMessage
      : null;

  switch (status) {
    case 400:
    case 422:
      return readable ?? l10n.errValidation;
    case 401:
      return l10n.errSessionExpired;
    case 403:
      return l10n.errForbidden;
    case 404:
      return l10n.errNotFound;
    case 409:
      return readable ?? l10n.errConflict;
    case 413:
      return l10n.errFileTooLarge;
    case 429:
      return l10n.errRateLimit;
  }
  if (status != null && status >= 500) return l10n.errServer;
  return readable ?? l10n.errGeneric;
}

final _cyrillic = RegExp('[а-яА-ЯёЁ]');

/// Английские тексты сервера («not found», «forbidden») — технические,
/// поэтому для английского интерфейса текст сервера не используем.
bool _matchesLanguage(String? message, String localeName) =>
    message != null && localeName == 'ru' && _cyrillic.hasMatch(message);

String? _serverMessage(Object? data) {
  // Формат ответа: { "error": { "code": "...", "message": "..." } }
  if (data is Map && data['error'] is Map) {
    final message = (data['error'] as Map)['message'];
    if (message is String) return message;
  }
  return null;
}
