import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/group.dart';
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
      body: _buildBody(state),
    );
  }

  Widget _buildBody(GroupsState state) {
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
      return const Center(child: Text('Нет групп'));
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
