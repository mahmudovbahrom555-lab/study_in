import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/localization/locale_provider.dart';
import 'core/localization/locale_resolver.dart';
import 'core/router/app_router.dart';
import 'generated/l10n/app_localizations.dart';
import 'core/theme/app_theme.dart';
import 'features/auth/presentation/providers/auth_provider.dart';

/// Корневой виджет приложения.
class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    final locale = ref.watch(localeProvider);
    ref.watch(languageSyncProvider);

    return MaterialApp.router(
      title: 'MaMu Learn',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light(),
      darkTheme: AppTheme.dark(),
      themeMode: ThemeMode.system,
      routerConfig: router,
      // Язык выбирается правилом из locale_resolver.dart, а не системой Flutter.
      locale: locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      // Список языков — appLanguages в locale_resolver.dart.
      supportedLocales: appLocales,
    );
  }
}
