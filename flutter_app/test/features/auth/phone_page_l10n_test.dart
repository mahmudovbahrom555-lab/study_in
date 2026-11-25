import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/localization/l10n.dart';
import 'package:mamu_learn/core/localization/locale_provider.dart';
import 'package:mamu_learn/core/localization/locale_resolver.dart';
import 'package:mamu_learn/features/auth/data/birth_year.dart';
import 'package:mamu_learn/features/auth/presentation/pages/phone_page.dart';

class _MemoryLanguageStore implements LanguageStore {
  _MemoryLanguageStore(this.value);

  String? value;

  @override
  String? read() => value;

  @override
  Future<void> write(String languageCode) async => value = languageCode;
}

class _MemoryBirthYearStore implements BirthYearStore {
  _MemoryBirthYearStore(this.value);

  int? value;

  @override
  int? read() => value;

  @override
  Future<void> write(int year) async => value = year;
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

Future<_MemoryLanguageStore> _pump(
  WidgetTester tester,
  String saved, {
  int? birthYear = 1990,
}) async {
  final store = _MemoryLanguageStore(saved);
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        localeProvider.overrideWith((ref) => LocaleController(store)),
        birthYearProvider.overrideWith(
          (ref) => BirthYearController(_MemoryBirthYearStore(birthYear)),
        ),
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

    await tester.tap(find.text('Русский')); // открыть меню языков
    await tester.pumpAndSettle();
    for (final l in appLanguages) {
      expect(find.text(l.nativeName), findsWidgets);
    }
    await tester.tap(find.text('English'));
    await tester.pumpAndSettle();

    expect(find.text('Sign in'), findsOneWidget);
    expect(store.value, 'en');
  });

  testWidgets('сначала спрашивает год рождения, потом вход', (tester) async {
    await _pump(tester, 'ru', birthYear: null);
    expect(find.text('Год рождения'), findsOneWidget);
    expect(find.text('Получить код'), findsNothing);

    await tester.tap(find.text('Выберите год'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('${DateTime.now().year - 3}').last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Продолжить'));
    await tester.pumpAndSettle();

    expect(find.text('Получить код'), findsOneWidget);
  });

  testWidgets('маленький экран и крупный шрифт — без переполнения', (
    tester,
  ) async {
    // Как телефон с открытой клавиатурой: по высоте остаётся мало места.
    tester.view.physicalSize = const Size(360, 420);
    tester.view.devicePixelRatio = 1;
    tester.platformDispatcher.textScaleFactorTestValue = 1.5;
    addTearDown(tester.view.reset);
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await _pump(tester, 'ru');

    // Переполнение («BOTTOM OVERFLOWED») Flutter сообщает как исключение.
    expect(tester.takeException(), isNull);
    expect(find.text('Получить код'), findsOneWidget);
  });

  test('младше 13 — по году рождения, с запасом в год', () {
    final now = DateTime(2026, 11, 24);
    expect(isChildBirthYear(2014, now: now), isTrue);
    expect(isChildBirthYear(2013, now: now), isTrue);
    expect(isChildBirthYear(2012, now: now), isFalse);
  });
}
