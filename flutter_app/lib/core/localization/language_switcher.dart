import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'locale_provider.dart';
import 'locale_resolver.dart';

/// Переключатель языка интерфейса.
///
/// Названия языков — на самих языках, чтобы человек нашёл свой язык,
/// даже если автоопределение выбрало непонятный ему.
class LanguageSwitcher extends ConsumerWidget {
  const LanguageSwitcher({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final current = ref.watch(localeProvider);

    return SegmentedButton<String>(
      showSelectedIcon: false,
      style: const ButtonStyle(visualDensity: VisualDensity.compact),
      segments: const [
        ButtonSegment(value: 'ru', label: Text('Русский')),
        ButtonSegment(value: 'en', label: Text('English')),
      ],
      selected: {current.languageCode},
      onSelectionChanged: (s) => ref
          .read(localeProvider.notifier)
          .select(s.first == 'en' ? enLocale : ruLocale),
    );
  }
}
