import 'package:package_info_plus/package_info_plus.dart';
import 'package:sentry_flutter/sentry_flutter.dart';

import '../config/app_config.dart';

/// Запускает приложение со сбором крашей и необработанных ошибок (Sentry).
///
/// Без `--dart-define=SENTRY_DSN=...` сбор выключен. Персональные данные
/// не отправляются (правило переноса данных, docs/architecture.md): на экранах
/// имена и оценки детей, поэтому без скриншотов и снимка интерфейса.
Future<void> runWithCrashReporting(PackageInfo info, void Function() appRunner) async {
  if (AppConfig.sentryDsn.isEmpty) {
    appRunner();
    return;
  }
  await SentryFlutter.init(
    (o) {
      o.dsn = AppConfig.sentryDsn;
      o.environment = AppConfig.env;
      o.release = '${info.packageName}@${info.version}+${info.buildNumber}';
      o.sendDefaultPii = false;
      o.attachScreenshot = false;
      o.attachViewHierarchy = false;
      o.enablePrintBreadcrumbs = false; // в логах бывают ответы сервера
      o.tracesSampleRate = 0; // скорость пока не меряем — только ошибки
    },
    appRunner: appRunner,
  );
}
