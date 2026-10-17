/// Конфигурация приложения. Значения задаются через --dart-define.
///
/// Запуск:
///   Android эмулятор: flutter run --dart-define=API_URL=http://10.0.2.2:8080/api/v1
///   iOS симулятор:    flutter run --dart-define=API_URL=http://localhost:8080/api/v1
///   Реальное устройство: flutter run --dart-define=API_URL=http://<IP>:8080/api/v1
class AppConfig {
  /// Базовый URL API. По умолчанию — Android-эмулятор (10.0.2.2 = хост-машина).
  /// Для iOS симулятора передайте --dart-define=API_URL=http://localhost:8080/api/v1
  static const String apiUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'http://10.0.2.2:8080/api/v1',
  );

  /// Окружение: development, staging, production.
  static const String env = String.fromEnvironment(
    'ENV',
    defaultValue: 'development',
  );

  /// Включён ли подробный логгинг.
  static bool get isDev => env == 'development';

  /// Куда слать краши (Sentry). Пусто — сбор выключен.
  static const String sentryDsn = String.fromEnvironment('SENTRY_DSN');

  /// Политика конфиденциальности (docs/legal/privacy-policy.ru.md). Страница
  /// публикуется на домене; до этого ссылка не откроется.
  static const String privacyUrl = String.fromEnvironment(
    'PRIVACY_URL',
    defaultValue: 'https://mamulearn.uz/privacy',
  );

  /// Версия политики, с которой человек соглашается на экране входа.
  /// Меняется вместе с текстом политики — сервер хранит её в журнале согласий.
  static const String privacyVersion = '2026-10-17';
}
