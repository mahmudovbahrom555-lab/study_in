import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../localization/l10n.dart';
import '../network/error_message.dart';

/// Единый экран ошибки: понятный текст и, если передан [onRetry], кнопка «Повторить».
///
/// Передайте либо [error] (исключение — будет переведено в текст),
/// либо готовое [message].
class ErrorView extends StatelessWidget {
  const ErrorView({super.key, this.error, this.message, this.onRetry})
      : assert(error != null || message != null);

  final Object? error;
  final String? message;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final text = message ?? userErrorMessage(error!);

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 420),
        child: Padding(
          padding: const EdgeInsets.all(24),
          // liveRegion: экранный диктор зачитает ошибку, когда она появится.
          child: Semantics(
            liveRegion: true,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                ExcludeSemantics(
                  child: Icon(
                    LucideIcons.circleAlert,
                    size: 48,
                    color: theme.colorScheme.error,
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  text,
                  textAlign: TextAlign.center,
                  style: theme.textTheme.bodyLarge,
                ),
                if (onRetry != null) ...[
                  const SizedBox(height: 16),
                  FilledButton.icon(
                    onPressed: onRetry,
                    icon: const Icon(LucideIcons.refreshCw),
                    label: Text(context.l10n.retry),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
