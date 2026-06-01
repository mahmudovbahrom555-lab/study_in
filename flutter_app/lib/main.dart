import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import 'app.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Инициализация локального хранилища.
  await Hive.initFlutter();

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
