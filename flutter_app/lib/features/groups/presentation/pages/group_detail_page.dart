import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/group.dart';
import '../providers/groups_provider.dart';
import '../widgets/member_tile.dart';

class GroupDetailPage extends ConsumerStatefulWidget {
  const GroupDetailPage({super.key, required this.groupId});

  final String groupId;

  @override
  ConsumerState<GroupDetailPage> createState() => _GroupDetailPageState();
}

class _GroupDetailPageState extends ConsumerState<GroupDetailPage> {
  @override
  Widget build(BuildContext context) {
    final groupsState = ref.watch(groupsProvider);
    final group = groupsState.groups
        .where((g) => g.id == widget.groupId)
        .firstOrNull;
    final membersAsync = ref.watch(membersProvider(widget.groupId));
    final authState = ref.watch(authProvider);
    final isTeacher = authState.user?.role == 'teacher';
    final isOwner = isTeacher && group?.teacherId == authState.user?.id;

    if (group == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(group.name),
        actions: [
          if (isOwner) ...[
            IconButton(
              icon: Icon(
                group.isArchived ? Icons.unarchive : Icons.archive,
              ),
              onPressed: () =>
                  ref.read(groupsProvider.notifier).archiveGroup(group.id),
            ),
            PopupMenuButton<String>(
              onSelected: (v) => _onMenuSelected(context, v, group),
              itemBuilder: (_) => [
                const PopupMenuItem(value: 'edit', child: Text('Редактировать')),
                const PopupMenuItem(
                  value: 'delete',
                  child: Text('Удалить', style: TextStyle(color: Colors.red)),
                ),
              ],
            ),
          ],
          if (!isTeacher)
            TextButton(
              onPressed: () => _confirmLeave(context, group),
              child: const Text('Покинуть'),
            ),
        ],
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _GroupInfoCard(group: group, isOwner: isOwner),
          const Divider(height: 1),
          _QuickActions(groupId: widget.groupId, isTeacher: isTeacher),
          const Divider(height: 1),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
            child: Text(
              'Участники',
              style: Theme.of(context).textTheme.titleMedium,
            ),
          ),
          Expanded(
            child: membersAsync.when(
              loading: () =>
                  const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text(e.toString())),
              data: (members) => members.isEmpty
                  ? const Center(child: Text('Нет участников'))
                  : ListView.builder(
                      itemCount: members.length,
                      itemBuilder: (context, i) => MemberTile(
                        member: members[i],
                        isOwner: isOwner,
                        onRemove: isOwner
                            ? () => _confirmRemove(
                                  context,
                                  members[i],
                                )
                            : null,
                        onPaymentTap: isOwner
                            ? (status) => ref
                                .read(membersProvider(widget.groupId).notifier)
                                .updatePayment(members[i].studentId, status)
                            : null,
                      ),
                    ),
            ),
          ),
        ],
      ),
    );
  }

  void _onMenuSelected(BuildContext context, String value, Group group) {
    if (value == 'delete') _confirmDelete(context, group);
  }

  Future<void> _confirmDelete(BuildContext context, Group group) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Удалить группу?'),
        content: Text('«${group.name}» будет удалена безвозвратно.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Отмена')),
          TextButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Удалить',
                  style: TextStyle(color: Colors.red))),
        ],
      ),
    );
    if (ok == true && mounted) {
      await ref.read(groupsProvider.notifier).deleteGroup(group.id);
      if (mounted) Navigator.of(context).pop();
    }
  }

  Future<void> _confirmLeave(BuildContext context, Group group) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Покинуть группу?'),
        content: Text('Вы покинете «${group.name}».'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Отмена')),
          TextButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Покинуть')),
        ],
      ),
    );
    if (ok == true && mounted) {
      await ref
          .read(groupsRepositoryProvider)
          .leaveGroup(group.id);
      await ref.read(groupsProvider.notifier).load();
      if (mounted) Navigator.of(context).pop();
    }
  }

  Future<void> _confirmRemove(BuildContext context, GroupMember member) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Удалить участника?'),
        content: Text('${member.name} будет удалён из группы.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Отмена')),
          TextButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Удалить',
                  style: TextStyle(color: Colors.red))),
        ],
      ),
    );
    if (ok == true) {
      await ref
          .read(membersProvider(widget.groupId).notifier)
          .removeMember(member.studentId);
    }
  }
}

// ─── Quick action buttons ─────────────────────────────────────────────────────

class _QuickActions extends StatelessWidget {
  const _QuickActions({required this.groupId, required this.isTeacher});

  final String groupId;
  final bool isTeacher;

  @override
  Widget build(BuildContext context) {
    final role = isTeacher ? 'teacher' : 'student';
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      child: Wrap(
        spacing: 8,
        children: [
          _ActionChip(
            icon: Icons.dynamic_feed,
            label: 'Лента',
            onTap: () => context.push(Routes.groupFeed(groupId)),
          ),
          _ActionChip(
            icon: Icons.assignment_outlined,
            label: 'Задания',
            onTap: () => context.push(Routes.groupAssignments(groupId)),
          ),
          _ActionChip(
            icon: Icons.quiz,
            label: 'Тесты',
            onTap: () => context.push(Routes.groupQuizzes(groupId)),
          ),
          _ActionChip(
            icon: Icons.grade,
            label: 'Оценки',
            onTap: () => context.push(
              '${Routes.groupGrades(groupId)}?role=$role',
            ),
          ),
          _ActionChip(
            icon: Icons.calendar_month,
            label: 'Посещ.',
            onTap: () => context.push(
              '${Routes.groupAttendance(groupId)}?role=$role',
            ),
          ),
          if (isTeacher)
            _ActionChip(
              icon: Icons.insights,
              label: 'AI Insights',
              onTap: () => context.push(Routes.aiInsights(groupId)),
            ),
        ],
      ),
    );
  }
}

class _ActionChip extends StatelessWidget {
  const _ActionChip({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ActionChip(
      avatar: Icon(icon, size: 18),
      label: Text(label),
      onPressed: onTap,
    );
  }
}

// ─── Group info card ──────────────────────────────────────────────────────────

class _GroupInfoCard extends StatelessWidget {
  const _GroupInfoCard({required this.group, required this.isOwner});

  final Group group;
  final bool isOwner;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (group.subject != null) ...[
            Text(group.subject!,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Theme.of(context).colorScheme.primary,
                    )),
            const SizedBox(height: 4),
          ],
          if (group.description != null) ...[
            Text(group.description!),
            const SizedBox(height: 8),
          ],
          if (isOwner)
            Row(
              children: [
                const Icon(Icons.link, size: 16),
                const SizedBox(width: 4),
                Text('Код: ${group.inviteCode}',
                    style: const TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
        ],
      ),
    );
  }
}
