import 'dart:ui';

/// Языки релиза (решение 2026-10-03, DECISIONS.md): русский и английский.
const ruLocale = Locale('ru');
const enLocale = Locale('en');
const appLocales = [ruLocale, enLocale];

/// Языки телефона, при которых по умолчанию включаем русский:
/// страны бывшего СССР, где русский широко понятен.
///
/// Намеренно НЕ включены украинский (uk), грузинский (ka), литовский (lt),
/// латышский (lv), эстонский (et): русский по умолчанию там многие воспримут
/// негативно — для них английский. Пользователь всегда может переключить язык.
const ruByDefaultLanguages = {
  'ru', // русский
  'uz', // узбекский
  'kk', // казахский
  'ky', // киргизский
  'tg', // таджикский
  'tk', // туркменский
  'az', // азербайджанский
  'hy', // армянский
  'be', // белорусский
};

/// Выбирает язык интерфейса.
///
/// 1. Язык, который пользователь уже выбрал в приложении ([savedLanguageCode]).
/// 2. Основной язык телефона — английский → английский (человек выбрал его сам,
///    даже если живёт в Узбекистане).
/// 3. Язык телефона из [ruByDefaultLanguages] → русский.
/// 4. Всё остальное (филиппинский и остальной мир) → английский.
Locale resolveAppLocale({
  String? savedLanguageCode,
  required List<Locale> deviceLocales,
}) {
  for (final l in appLocales) {
    if (l.languageCode == savedLanguageCode) return l;
  }

  // Решает основной язык телефона — тот, что пользователь поставил первым.
  final primary = deviceLocales.isEmpty ? null : deviceLocales.first.languageCode;
  if (primary == 'en') return enLocale;
  if (ruByDefaultLanguages.contains(primary)) return ruLocale;
  return enLocale;
}
