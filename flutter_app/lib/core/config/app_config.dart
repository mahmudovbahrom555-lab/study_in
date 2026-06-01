/// Конфигурация приложения. Значения задаются через --dart-define.
///
/// Запуск:
///   flutter run --dart-define=API_URL=http://localhost:8080
class AppConfig {
  /// Базовый URL API. По умолчанию — локальный сервер.
  static const String apiUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'http://localhost:8080/api/v1',
  );

  /// Окружение: development, staging, production.
  static const String env = String.fromEnvironment(
    'ENV',
    defaultValue: 'development',
  );

  /// Включён ли подробный логгинг.
  static bool get isDev => env == 'development';

  /// Версия приложения.
  static const String version = '0.1.0';
}
