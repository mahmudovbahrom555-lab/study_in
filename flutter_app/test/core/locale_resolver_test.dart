import 'dart:ui';

import 'package:flutter_test/flutter_test.dart';
import 'package:repetapp/core/localization/locale_resolver.dart';

String resolve(List<String> device, {String? saved}) => resolveAppLocale(
      savedLanguageCode: saved,
      deviceLocales: device.map(Locale.new).toList(),
    ).languageCode;

void main() {
  group('resolveAppLocale', () {
    test('ручной выбор пользователя главнее языка телефона', () {
      expect(resolve(['uz'], saved: 'en'), 'en');
      expect(resolve(['fil'], saved: 'ru'), 'ru');
    });

    test('сохранённый неподдерживаемый язык игнорируется', () {
      expect(resolve(['uz'], saved: 'uz'), 'ru');
    });

    test('телефон на русском или английском — этот язык', () {
      expect(resolve(['ru']), 'ru');
      expect(resolve(['en']), 'en');
    });

    test('страны бывшего СССР → русский', () {
      for (final lang in ['uz', 'kk', 'ky', 'tg', 'tk', 'az', 'hy', 'be']) {
        expect(resolve([lang]), 'ru', reason: lang);
      }
    });

    test('украинский, грузинский, Прибалтика → английский', () {
      for (final lang in ['uk', 'ka', 'lt', 'lv', 'et']) {
        expect(resolve([lang]), 'en', reason: lang);
      }
    });

    test('Филиппины и остальной мир → английский', () {
      for (final lang in ['fil', 'tl', 'hi', 'ar', 'es', 'zh']) {
        expect(resolve([lang]), 'en', reason: lang);
      }
    });

    test('решает основной язык телефона, а не второй в списке', () {
      expect(resolve(['en', 'uz']), 'en');
      expect(resolve(['uz', 'en']), 'ru');
    });

    test('пустой список языков → английский', () {
      expect(resolve([]), 'en');
    });
  });
}
