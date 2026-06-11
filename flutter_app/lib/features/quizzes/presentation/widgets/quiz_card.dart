import 'package:flutter/material.dart';

import '../../domain/entities/quiz.dart';

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
                          icon: Icons.repeat,
                          label: '${quiz.maxAttempts} поп.',
                        ),
                        if (quiz.timeLimit != null) ...[
                          const SizedBox(width: 8),
                          _Chip(
                            icon: Icons.timer_outlined,
                            label: '${quiz.timeLimit} мин',
                          ),
                        ],
                        const SizedBox(width: 8),
                        _Chip(
                          icon: quiz.isPublished
                              ? Icons.visibility
                              : Icons.visibility_off,
                          label: quiz.isPublished ? 'Опубл.' : 'Черновик',
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
                    quiz.isPublished ? Icons.unpublished : Icons.publish,
                    color: quiz.isPublished ? Colors.orange : Colors.green,
                  ),
                  tooltip: quiz.isPublished ? 'Снять' : 'Опубликовать',
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
        Text(label,
            style: TextStyle(fontSize: 12, color: c)),
      ],
    );
  }
}
