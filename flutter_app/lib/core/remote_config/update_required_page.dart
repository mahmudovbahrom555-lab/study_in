import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../localization/l10n.dart';
import 'remote_config_provider.dart';

/// Версия приложения ниже минимальной с сервера: дальше не пускаем —
/// старый клиент может испортить данные или упасть на новом API.
class UpdateRequiredPage extends ConsumerWidget {
  const UpdateRequiredPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = context.l10n;
    final theme = Theme.of(context);
    final packageName = ref.watch(appInfoProvider).packageName;

    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(LucideIcons.circleArrowUp, size: 64, color: theme.colorScheme.primary),
              const SizedBox(height: 24),
              Text(
                l10n.updateRequiredTitle,
                style: theme.textTheme.headlineSmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              Text(
                l10n.updateRequiredBody,
                style: theme.textTheme.bodyMedium,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),
              FilledButton(
                onPressed: () => launchUrl(
                  Uri.parse('https://play.google.com/store/apps/details?id=$packageName'),
                  mode: LaunchMode.externalApplication,
                ),
                child: Text(l10n.updateAction),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
