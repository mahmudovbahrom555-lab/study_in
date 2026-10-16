import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import 'l10n.dart';
import 'locale_provider.dart';
import 'locale_resolver.dart';

/// Переключатель языка интерфейса: текущий язык и меню со всеми языками
/// из [appLanguages] — вмещает любое их число.
class LanguageSwitcher extends ConsumerWidget {
  const LanguageSwitcher({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final code = ref.watch(localeProvider).languageCode;
    final current = appLanguages.firstWhere(
      (l) => l.locale.languageCode == code,
      orElse: () => appLanguages.first,
    );

    return PopupMenuButton<AppLanguage>(
      tooltip: context.l10n.language,
      initialValue: current,
      onSelected: (l) => ref.read(localeProvider.notifier).select(l.locale),
      itemBuilder: (_) => [
        for (final l in appLanguages)
          PopupMenuItem(value: l, child: Text(l.nativeName)),
      ],
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 12),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(current.nativeName),
            const SizedBox(width: 4),
            const Icon(LucideIcons.chevronDown, size: 18),
          ],
        ),
      ),
    );
  }
}
