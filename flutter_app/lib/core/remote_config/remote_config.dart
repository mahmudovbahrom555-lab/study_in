/// Ключи фич, которые сервер включает без релиза в стор (APP_FEATURES).
///
/// Новая фича: ключ здесь + проверка `isEnabled` там, где она появляется
/// в интерфейсе. Пока ключа нет в APP_FEATURES — фича скрыта.
abstract final class Feature {
  /// Тесты — скрыты в MVP до создания теста ИИ (DECISIONS.md, 2026-10-15).
  static const quizzes = 'quizzes';
}

/// Настройки с сервера (`GET /app-config`): минимальная версия и флаги фич.
class RemoteConfig {
  const RemoteConfig({this.minVersion, this.features = const {}});

  factory RemoteConfig.fromJson(Map<String, dynamic> json) => RemoteConfig(
        minVersion: json['min_version'] as String?,
        features: {
          for (final e in (json['features'] as Map? ?? const {}).entries)
            if (e.value == true) e.key as String,
        },
      );

  Map<String, dynamic> toJson() => {
        'min_version': minVersion,
        'features': {for (final f in features) f: true},
      };

  final String? minVersion;
  final Set<String> features;

  bool isEnabled(String feature) => features.contains(feature);

  /// Версия [appVersion] ниже минимальной — работать нельзя, нужно обновиться.
  bool requiresUpdate(String appVersion) {
    final min = minVersion;
    return min != null && compareVersions(appVersion, min) < 0;
  }
}

/// Сравнивает версии вида 1.2.3 (суффиксы +build и -pre не учитываются).
int compareVersions(String a, String b) {
  List<int> parts(String v) => v
      .split(RegExp('[+-]'))
      .first
      .split('.')
      .map((p) => int.tryParse(p) ?? 0)
      .toList();
  final pa = parts(a), pb = parts(b);
  for (var i = 0; i < 3; i++) {
    final x = i < pa.length ? pa[i] : 0;
    final y = i < pb.length ? pb[i] : 0;
    if (x != y) return x.compareTo(y);
  }
  return 0;
}
