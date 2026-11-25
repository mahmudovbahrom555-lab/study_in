import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../feed/domain/entities/post.dart';
import '../../../../core/localization/l10n.dart';

class PostCard extends StatelessWidget {
  const PostCard({
    super.key,
    required this.post,
    required this.isAuthor,
    required this.isTeacher,
    this.onDelete,
    this.onPin,
    this.onReact,
  });

  final Post post;
  final bool isAuthor;
  final bool isTeacher;
  final VoidCallback? onDelete;
  final VoidCallback? onPin;

  /// Только у ученика: нажатие на «Понятно» / «Есть вопрос».
  final ValueChanged<String>? onReact;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      color: post.pinned
          ? theme.colorScheme.primaryContainer.withValues(alpha: 0.3)
          : null,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                if (post.pinned) ...[
                  Icon(
                    LucideIcons.pin,
                    size: 14,
                    color: theme.colorScheme.primary,
                  ),
                  const SizedBox(width: 4),
                ],
                Text(
                  post.authorName,
                  style: theme.textTheme.labelMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const Spacer(),
                Text(
                  _formatDate(context.l10n, post.createdAt),
                  style: theme.textTheme.labelSmall,
                ),
                if (isTeacher || isAuthor)
                  PopupMenuButton<String>(
                    iconSize: 18,
                    onSelected: (v) {
                      if (v == 'delete') onDelete?.call();
                      if (v == 'pin') onPin?.call();
                    },
                    itemBuilder: (_) => [
                      if (isTeacher)
                        PopupMenuItem(
                          value: 'pin',
                          child: Text(
                            post.pinned ? context.l10n.unpin : context.l10n.pin,
                          ),
                        ),
                      if (isTeacher || isAuthor)
                        PopupMenuItem(
                          value: 'delete',
                          child: Text(
                            context.l10n.delete,
                            style: const TextStyle(color: Colors.red),
                          ),
                        ),
                    ],
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Text(post.body),
            const SizedBox(height: 8),
            _ReactionsRow(
              reactions: post.reactions,
              isTeacher: isTeacher,
              onReact: onReact,
            ),
          ],
        ),
      ),
    );
  }

  String _formatDate(AppLocalizations l10n, DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return l10n.justNow;
    if (diff.inHours < 1) return l10n.minutesAgo(diff.inMinutes);
    if (diff.inDays < 1) return l10n.hoursAgo(diff.inHours);
    return '${dt.day}.${dt.month.toString().padLeft(2, '0')}.${dt.year}';
  }
}

/// «Понятно» / «Есть вопрос»: ученик нажимает, репетитор смотрит, кто отметил.
class _ReactionsRow extends StatelessWidget {
  const _ReactionsRow({
    required this.reactions,
    required this.isTeacher,
    this.onReact,
  });

  final PostReactions reactions;
  final bool isTeacher;
  final ValueChanged<String>? onReact;

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Wrap(
      spacing: 8,
      children: [
        _chip(
          context,
          PostReactions.understoodKind,
          '✅ ${l10n.reactionUnderstood}',
          reactions.understood,
          reactions.understoodBy,
        ),
        _chip(
          context,
          PostReactions.questionKind,
          '❓ ${l10n.reactionQuestion}',
          reactions.question,
          reactions.questionBy,
        ),
      ],
    );
  }

  Widget _chip(
    BuildContext context,
    String kind,
    String label,
    int count,
    List<String> names,
  ) {
    final VoidCallback? onTap = onReact != null
        ? () => onReact!(kind)
        : (isTeacher && names.isNotEmpty)
        ? () => _showNames(context, label, names)
        : null;
    return FilterChip(
      label: Text(count > 0 ? '$label · $count' : label),
      selected: reactions.mine == kind,
      showCheckmark: false,
      onSelected: onTap == null ? null : (_) => onTap(),
    );
  }

  void _showNames(BuildContext context, String title, List<String> names) {
    showModalBottomSheet<void>(
      context: context,
      builder: (context) => SafeArea(
        child: ListView(
          shrinkWrap: true,
          padding: const EdgeInsets.all(16),
          children: [
            Text(title, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            for (final name in names) ListTile(title: Text(name)),
          ],
        ),
      ),
    );
  }
}
