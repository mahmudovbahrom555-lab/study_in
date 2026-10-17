import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:package_info_plus/package_info_plus.dart';

import 'app.dart';
import 'core/app_restart.dart';
import 'core/localization/locale_provider.dart';
import 'core/monitoring/crash_reporting.dart';
import 'core/remote_config/remote_config_provider.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Инициализация локального хранилища.
  await Hive.initFlutter();
  await Hive.openBox<dynamic>(settingsBoxName);
  final appInfo = await PackageInfo.fromPlatform();

  // На последующих этапах здесь будет:
  // - Инициализация Firebase (push)
  // - Регистрация Hive адаптеров

  await runWithCrashReporting(
    appInfo,
    () => runApp(
      // AppRestartScope снаружи ProviderScope: при выходе из аккаунта
      // пересоздаются все провайдеры (см. AccountPage).
      AppRestartScope(
        child: ProviderScope(
          overrides: [appInfoProvider.overrideWithValue(appInfo)],
          child: const App(),
        ),
      ),
    ),
  );
}
