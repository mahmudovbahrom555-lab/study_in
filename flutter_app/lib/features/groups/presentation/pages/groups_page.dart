import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../ai/presentation/providers/ai_insights_provider.dart';
import '../providers/groups_provider.dart';
import '../widgets/group_card.dart';
import '../widgets/create_group_sheet.dart';
import '../widgets/join_group_sheet.dart';

class GroupsPage extends ConsumerStatefulWidget {
  const GroupsPage({super.key});

  @override
  ConsumerState<GroupsPage> createState() => _GroupsPageState();
}

class _GroupsPageState extends ConsumerState<GroupsPage> {
  bool _demoLoading = false;

  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(groupsProvider.notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(groupsProvider);
    final role = ref.watch(authProvider).user?.role;
    final isTeacher = role == 'teacher';

    return Scaffold(
      appBar: AppBar(
        title: const Text('Мои группы'),
        actions: [
          if (!isTeacher)
            IconButton(
              icon: const Icon(Icons.add_link),
              tooltip: 'Вступить по коду',
              onPressed: () => _showJoinSheet(context),
            ),
        ],
      ),
      floatingActionButton: isTeacher
          ? FloatingActionButton(
              onPressed: () => _showCreateSheet(context),
              child: const Icon(Icons.add),
            )
          : null,
      body: _buildBody(state, isTeacher),
    );
  }

  Widget _buildBody(GroupsState state, bool isTeacher) {
    if (state.isLoading && state.groups.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state.error != null && state.groups.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(state.error!, style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: () => ref.read(groupsProvider.notifier).load(),
              child: const Text('Повторить'),
            ),
          ],
        ),
      );
    }
    if (state.groups.isEmpty) {
      return isTeacher
          ? _TeacherEmptyState(onDemo: _openDemo, demoLoading: _demoLoading)
          : _StudentEmptyState(onJoin: () => _showJoinSheet(context));
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(groupsProvider.notifier).load(),
      child: ListView.separated(
        padding: const EdgeInsets.all(16),
        itemCount: state.groups.length,
        separatorBuilder: (_, __) => const SizedBox(height: 8),
        itemBuilder: (context, i) {
          final g = state.groups[i];
          return GroupCard(
            group: g,
            onTap: () => context.push(Routes.group(g.id)),
          );
        },
      ),
    );
  }

  Future<void> _openDemo() async {
    setState(() => _demoLoading = true);
    try {
      final groupId =
          await ref.read(aiRepositoryProvider).createDemoGroup();
      if (mounted) context.push('/groups/$groupId/ai-insights');
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Не удалось открыть демо')),
        );
      }
    } finally {
      if (mounted) setState(() => _demoLoading = false);
    }
  }

  void _showCreateSheet(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => const CreateGroupSheet(),
    );
  }

  void _showJoinSheet(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => const JoinGroupSheet(),
    );
  }
}

class _TeacherEmptyState extends StatelessWidget {
  const _TeacherEmptyState({required this.onDemo, this.demoLoading = false});

  final VoidCallback onDemo;
  final bool demoLoading;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.school_outlined,
              size: 72,
              color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.4),
            ),
            const SizedBox(height: 20),
            Text(
              'Добро пожаловать!',
              style: Theme.of(context).textTheme.headlineSmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'Создайте первую группу или посмотрите как работает платформа на демо-данных.',
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: Colors.grey.shade600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            FilledButton.icon(
              icon: const Icon(Icons.add),
              label: const Text('Создать первую группу'),
              onPressed: () => showModalBottomSheet<void>(
                context: context,
                isScrollControlled: true,
                builder: (_) => const CreateGroupSheet(),
              ),
            ),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              icon: demoLoading
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.play_circle_outline),
              label: const Text('Посмотреть демо'),
              onPressed: demoLoading ? null : onDemo,
            ),
          ],
        ),
      ),
    );
  }
}

class _StudentEmptyState extends StatelessWidget {
  const _StudentEmptyState({required this.onJoin});

  final VoidCallback onJoin;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.groups_outlined,
              size: 72,
              color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.4),
            ),
            const SizedBox(height: 20),
            Text(
              'Вы ещё не в группах',
              style: Theme.of(context).textTheme.headlineSmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'Попросите учителя поделиться кодом группы и вступите по нему.',
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: Colors.grey.shade600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            FilledButton.icon(
              icon: const Icon(Icons.add_link),
              label: const Text('Вступить по коду'),
              onPressed: onJoin,
            ),
          ],
        ),
      ),
    );
  }
}
