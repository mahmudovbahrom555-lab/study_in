import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/localization/l10n.dart';
import 'package:mamu_learn/core/localization/locale_provider.dart';
import 'package:mamu_learn/core/localization/locale_resolver.dart';
import 'package:mamu_learn/features/auth/presentation/pages/phone_page.dart';

class _MemoryLanguageStore implements LanguageStore {
  _MemoryLanguageStore(this.value);

  String? value;

  @override
  String? read() => value;

  @override
  Future<void> write(String languageCode) async => value = languageCode;
}

/// Мини-приложение: язык MaterialApp берётся из localeProvider, как в App.
class _TestApp extends ConsumerWidget {
  const _TestApp();

  @override
  Widget build(BuildContext context, WidgetRef ref) => MaterialApp(
        locale: ref.watch(localeProvider),
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: appLocales,
        home: const PhonePage(),
      );
}

Future<_MemoryLanguageStore> _pump(WidgetTester tester, String saved) async {
  final store = _MemoryLanguageStore(saved);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        localeProvider.overrideWith((ref) => LocaleController(store)),
      ],
      child: const _TestApp(),
    ),
  );
  await tester.pumpAndSettle();
  return store;
}

void main() {
  testWidgets('экран входа на русском', (tester) async {
    await _pump(tester, 'ru');
    expect(find.text('Вход'), findsOneWidget);
    expect(find.text('Получить код'), findsOneWidget);
  });

  testWidgets('экран входа на английском', (tester) async {
    await _pump(tester, 'en');
    expect(find.text('Sign in'), findsOneWidget);
    expect(find.text('Get code'), findsOneWidget);
  });

  testWidgets('переключатель меняет язык и сохраняет выбор', (tester) async {
    final store = await _pump(tester, 'ru');

    await tester.tap(find.text('English'));
    await tester.pumpAndSettle();

    expect(find.text('Sign in'), findsOneWidget);
    expect(store.value, 'en');
  });
}
