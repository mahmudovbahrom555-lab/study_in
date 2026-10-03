import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import 'app.dart';
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
    const ProviderScope(
      child: App(),
    ),
  );
}
