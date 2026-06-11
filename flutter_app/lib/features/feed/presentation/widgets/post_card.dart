import 'package:flutter/material.dart';

import '../../../feed/domain/entities/post.dart';

class PostCard extends StatelessWidget {
  const PostCard({
    super.key,
    required this.post,
    required this.isAuthor,
    required this.isTeacher,
    this.onDelete,
    this.onPin,
  });

  final Post post;
  final bool isAuthor;
  final bool isTeacher;
  final VoidCallback? onDelete;
  final VoidCallback? onPin;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      color: post.pinned
          ? theme.colorScheme.primaryContainer.withOpacity(0.3)
          : null,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                if (post.pinned) ...[
                  Icon(Icons.push_pin,
                      size: 14, color: theme.colorScheme.primary),
                  const SizedBox(width: 4),
                ],
                Text(
                  post.authorName,
                  style: theme.textTheme.labelMedium
                      ?.copyWith(fontWeight: FontWeight.bold),
                ),
                const Spacer(),
                Text(
                  _formatDate(post.createdAt),
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
                          child:
                              Text(post.pinned ? 'Открепить' : 'Закрепить'),
                        ),
                      if (isTeacher || isAuthor)
                        const PopupMenuItem(
                          value: 'delete',
                          child: Text('Удалить',
                              style: TextStyle(color: Colors.red)),
                        ),
                    ],
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Text(post.body),
            if (post.attachments.isNotEmpty) ...[
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: post.attachments
                    .map((a) => _AttachmentChip(attachment: a))
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }

  String _formatDate(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return 'только что';
    if (diff.inHours < 1) return '${diff.inMinutes} мин назад';
    if (diff.inDays < 1) return '${diff.inHours} ч назад';
    return '${dt.day}.${dt.month.toString().padLeft(2, '0')}.${dt.year}';
  }
}

class _AttachmentChip extends StatelessWidget {
  const _AttachmentChip({required this.attachment});

  final PostAttachment attachment;

  @override
  Widget build(BuildContext context) {
    return ActionChip(
      avatar: const Icon(Icons.attach_file, size: 14),
      label: Text(
        attachment.filename,
        overflow: TextOverflow.ellipsis,
        maxLines: 1,
      ),
      onPressed: () {
        // In Etap 4, this will open/download the file via signed URL.
      },
    );
  }
}
