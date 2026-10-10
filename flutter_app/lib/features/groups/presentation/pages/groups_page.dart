import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/router/routes.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../ai/presentation/providers/ai_insights_provider.dart';
import '../providers/groups_provider.dart';
import '../widgets/group_card.dart';
import '../widgets/group_form_sheet.dart';
import '../widgets/join_group_sheet.dart';
import '../../../../core/widgets/error_view.dart';
import '../../../../core/localization/l10n.dart';

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
        title: Text(context.l10n.myGroups),
        actions: [
          if (!isTeacher)
            IconButton(
              icon: const Icon(LucideIcons.link),
              tooltip: context.l10n.joinByCode,
              onPressed: () => _showJoinSheet(context),
            ),
        ],
      ),
      floatingActionButton: isTeacher
          ? FloatingActionButton(
            tooltip: context.l10n.createGroup,
              onPressed: () => _showCreateSheet(context),
              child: const Icon(LucideIcons.plus),
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
      return ErrorView(
        message: state.error!,
        onRetry: () => ref.read(groupsProvider.notifier).load(),
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
          SnackBar(content: Text(context.l10n.demoOpenFailed)),
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
      builder: (_) => const GroupFormSheet(),
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
              LucideIcons.presentation,
              size: 72,
              color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.4),
            ),
            const SizedBox(height: 20),
            Text(
              context.l10n.welcomeTitle,
              style: Theme.of(context).textTheme.headlineSmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              context.l10n.welcomeTeacherBody,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: Colors.grey.shade600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            FilledButton.icon(
              icon: const Icon(LucideIcons.plus),
              label: Text(context.l10n.createFirstGroup),
              onPressed: () => showModalBottomSheet<void>(
                context: context,
                isScrollControlled: true,
                builder: (_) => const GroupFormSheet(),
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
                  : const Icon(LucideIcons.circlePlay),
              label: Text(context.l10n.viewDemo),
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
              LucideIcons.users,
              size: 72,
              color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.4),
            ),
            const SizedBox(height: 20),
            Text(
              context.l10n.noGroupsStudentTitle,
              style: Theme.of(context).textTheme.headlineSmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              context.l10n.noGroupsStudentBody,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: Colors.grey.shade600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            FilledButton.icon(
              icon: const Icon(LucideIcons.link),
              label: Text(context.l10n.joinByCode),
              onPressed: onJoin,
            ),
          ],
        ),
      ),
    );
  }
}
