import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../domain/entities/quiz.dart';
import '../../../../core/localization/l10n.dart';

class QuizCard extends StatelessWidget {
  const QuizCard({
    super.key,
    required this.quiz,
    required this.onTap,
    this.onTogglePublish,
  });

  final Quiz quiz;
  final VoidCallback onTap;
  final VoidCallback? onTogglePublish;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      quiz.title,
                      style: theme.textTheme.titleMedium,
                    ),
                    if (quiz.description != null) ...[
                      const SizedBox(height: 4),
                      Text(
                        quiz.description!,
                        style: theme.textTheme.bodySmall,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        _Chip(
                          icon: LucideIcons.repeat,
                          label: context.l10n.attemptsShortChip(quiz.maxAttempts),
                        ),
                        if (quiz.timeLimit != null) ...[
                          const SizedBox(width: 8),
                          _Chip(
                            icon: LucideIcons.timer,
                            label: context.l10n.minutesShort(quiz.timeLimit!),
                          ),
                        ],
                        const SizedBox(width: 8),
                        if (quiz.isClosed)
                          _Chip(
                            icon: LucideIcons.lock,
                            label: context.l10n.closed,
                            color: Colors.grey,
                          )
                        else if (quiz.isNotYetOpen)
                          _Chip(
                            icon: LucideIcons.clock,
                            label: context.l10n.notStarted,
                            color: Colors.blue,
                          )
                        else
                          _Chip(
                            icon: quiz.isPublished
                                ? LucideIcons.eye
                                : LucideIcons.eyeOff,
                            label: quiz.isPublished ? context.l10n.publishedShort : context.l10n.draft,
                            color: quiz.isPublished
                                ? Colors.green
                                : Colors.orange,
                          ),
                      ],
                    ),
                  ],
                ),
              ),
              if (onTogglePublish != null)
                IconButton(
                  icon: Icon(
                    quiz.isPublished ? LucideIcons.eyeOff : LucideIcons.send,
                    color: quiz.isPublished ? Colors.orange : Colors.green,
                  ),
                  tooltip: quiz.isPublished ? context.l10n.unpublish : context.l10n.publish,
                  onPressed: onTogglePublish,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({required this.icon, required this.label, this.color});

  final IconData icon;
  final String label;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final c = color ?? Theme.of(context).colorScheme.primary;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: c),
        const SizedBox(width: 4),
        Text(
          label,
          style: TextStyle(fontSize: 12, color: c),
        ),
      ],
    );
  }
}
