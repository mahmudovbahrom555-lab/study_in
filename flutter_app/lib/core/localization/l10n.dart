import 'package:flutter/widgets.dart';

import '../../generated/l10n/app_localizations.dart';
import 'locale_resolver.dart';

export '../../generated/l10n/app_localizations.dart';

/// Короткий доступ к переводам в виджетах: `context.l10n.retry`.
extension L10nContext on BuildContext {
  AppLocalizations get l10n => AppLocalizations.of(this);
}

/// Текущий язык интерфейса для кода без BuildContext (провайдеры, сервисы).
/// Обновляет [LocaleController] при каждой смене языка.
Locale currentAppLocale = enLocale;

/// Переводы для кода без BuildContext — например, текст ошибки в провайдере.
AppLocalizations get currentL10n => lookupAppLocalizations(currentAppLocale);
