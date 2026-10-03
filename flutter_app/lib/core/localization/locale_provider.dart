import 'dart:ui';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import 'l10n.dart';
import 'locale_resolver.dart';

/// Hive-бокс с настройками устройства. Открывается в main() до runApp.
const settingsBoxName = 'settings';

/// Где хранится выбранный пользователем язык.
abstract interface class LanguageStore {
  String? read();
  Future<void> write(String languageCode);
}

class HiveLanguageStore implements LanguageStore {
  HiveLanguageStore(this._box);

  static const _key = 'language';
  final Box<dynamic> _box;

  @override
  String? read() => _box.get(_key) as String?;

  @override
  Future<void> write(String languageCode) => _box.put(_key, languageCode);
}

/// Текущий язык интерфейса. При первом запуске выбирается по языку телефона
/// ([resolveAppLocale]); после ручного выбора — сохраняется на устройстве.
class LocaleController extends StateNotifier<Locale> {
  LocaleController(this._store, {List<Locale>? deviceLocales})
      : super(
          resolveAppLocale(
            savedLanguageCode: _store.read(),
            deviceLocales: deviceLocales ?? PlatformDispatcher.instance.locales,
          ),
        ) {
    currentAppLocale = state;
  }

  final LanguageStore _store;

  /// Ручной выбор языка пользователем — главнее любого автоопределения.
  Future<void> select(Locale locale) async {
    currentAppLocale = locale;
    state = locale;
    await _store.write(locale.languageCode);
  }
}

final localeProvider = StateNotifierProvider<LocaleController, Locale>(
  (ref) => LocaleController(
    HiveLanguageStore(Hive.box<dynamic>(settingsBoxName)),
  ),
);
