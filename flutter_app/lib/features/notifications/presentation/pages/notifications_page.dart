import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/notification.dart';
import '../providers/notifications_provider.dart';
import '../../../../core/localization/l10n.dart';

class NotificationsPage extends ConsumerStatefulWidget {
  const NotificationsPage({super.key});

  @override
  ConsumerState<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends ConsumerState<NotificationsPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(notificationsProvider.notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(notificationsProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(context.l10n.notifications),
        actions: [
          if (state.unreadCount > 0)
            TextButton(
              onPressed: () =>
                  ref.read(notificationsProvider.notifier).markAllRead(),
              child: Text(context.l10n.readAll),
            ),
        ],
      ),
      body: _buildBody(state),
    );
  }

  Widget _buildBody(NotificationsState state) {
    if (state.isLoading && state.notifications.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state.notifications.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.notifications_none_outlined, size: 64, color: Colors.grey),
            const SizedBox(height: 12),
            Text(context.l10n.noNotifications, style: const TextStyle(color: Colors.grey)),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(notificationsProvider.notifier).load(),
      child: ListView.separated(
        itemCount: state.notifications.length,
        separatorBuilder: (_, __) => const Divider(height: 1),
        itemBuilder: (context, i) {
          final n = state.notifications[i];
          return _NotificationTile(
            notification: n,
            onTap: () {
              if (!n.isRead) {
                ref.read(notificationsProvider.notifier).markRead(n.id);
              }
            },
          );
        },
      ),
    );
  }
}

class _NotificationTile extends StatelessWidget {
  const _NotificationTile({required this.notification, required this.onTap});

  final AppNotification notification;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return ListTile(
      tileColor: notification.isRead
          ? null
          : colorScheme.primaryContainer.withValues(alpha: 0.15),
      leading: CircleAvatar(
        backgroundColor: colorScheme.secondaryContainer,
        child: Icon(_iconForType(notification.type), size: 20),
      ),
      title: Text(
        notification.title,
        style: TextStyle(
          fontWeight:
              notification.isRead ? FontWeight.normal : FontWeight.bold,
        ),
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(notification.body, maxLines: 2, overflow: TextOverflow.ellipsis),
          const SizedBox(height: 2),
          Text(
            _formatAge(context.l10n, notification.createdAt),
            style: const TextStyle(fontSize: 11, color: Colors.grey),
          ),
        ],
      ),
      isThreeLine: true,
      onTap: onTap,
    );
  }

  String _formatAge(AppLocalizations l10n, DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 1) return l10n.justNow;
    if (diff.inHours < 1) return l10n.minutesAgo(diff.inMinutes);
    if (diff.inDays < 1) return l10n.hoursAgo(diff.inHours);
    if (diff.inDays < 7) return l10n.daysAgo(diff.inDays);
    return '${dt.day}.${dt.month.toString().padLeft(2, '0')}.${dt.year}';
  }

  IconData _iconForType(String? type) => switch (type) {
        'quiz_published' => Icons.quiz_outlined,
        'assignment_due' => Icons.assignment_outlined,
        'grade_posted' => Icons.grade_outlined,
        'at_risk' => Icons.warning_amber_outlined,
        'celebrate' => Icons.celebration_outlined,
        _ => Icons.notifications_outlined,
      };
}
