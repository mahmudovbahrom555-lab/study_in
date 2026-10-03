import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import 'app.dart';
import 'core/app_restart.dart';
import 'core/localization/locale_provider.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Инициализация локального хранилища.
  await Hive.initFlutter();
  await Hive.openBox<dynamic>(settingsBoxName);

  // На последующих этапах здесь будет:
  // - Инициализация Firebase (push)
  // - Регистрация Hive адаптеров
  // - Инициализация Sentry

  runApp(
    // AppRestartScope снаружи ProviderScope: при выходе из аккаунта
    // пересоздаются все провайдеры (см. AccountPage).
    const AppRestartScope(
      child: ProviderScope(
        child: App(),
      ),
    ),
  );
}
