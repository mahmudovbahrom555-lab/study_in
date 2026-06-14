import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../attendance/presentation/pages/attendance_page.dart';
import '../../../grades/presentation/pages/grades_page.dart';
import '../../../reports/presentation/pages/parent_roi_page.dart';
import '../../domain/entities/parent_link.dart';
import '../providers/parents_provider.dart';

class ParentsPage extends ConsumerStatefulWidget {
  const ParentsPage({super.key});

  @override
  ConsumerState<ParentsPage> createState() => _ParentsPageState();
}

class _ParentsPageState extends ConsumerState<ParentsPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(childrenProvider.notifier).load(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(childrenProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Мои дети')),
      floatingActionButton: FloatingActionButton.extended(
        icon: const Icon(Icons.link),
        label: const Text('Добавить'),
        onPressed: () => _showLinkDialog(context),
      ),
      body: Builder(
        builder: (_) {
          if (state.isLoading && state.children.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state.error != null && state.children.isEmpty) {
            return Center(child: Text('Ошибка: ${state.error}'));
          }
          if (state.children.isEmpty) {
            return const Center(
              child: Text('Нет привязанных детей.\nНажмите + чтобы добавить.'),
            );
          }
          return RefreshIndicator(
            onRefresh: () => ref.read(childrenProvider.notifier).load(),
            child: ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: state.children.length,
              itemBuilder: (_, i) => _ChildCard(link: state.children[i]),
            ),
          );
        },
      ),
    );
  }

  void _showLinkDialog(BuildContext context) {
    final ctrl = TextEditingController();
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Привязать ребёнка'),
        content: TextField(
          controller: ctrl,
          decoration: const InputDecoration(
            labelText: 'ID ученика',
            hintText: 'UUID ученика',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Отмена'),
          ),
          FilledButton(
            onPressed: () {
              final id = ctrl.text.trim();
              if (id.isNotEmpty) {
                ref.read(childrenProvider.notifier).linkChild(id);
              }
              Navigator.pop(ctx);
            },
            child: const Text('Привязать'),
          ),
        ],
      ),
    );
  }
}

// ─── Child card ───────────────────────────────────────────────────────────────

class _ChildCard extends ConsumerWidget {
  const _ChildCard({required this.link});

  final ParentLink link;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ExpansionTile(
        leading: const CircleAvatar(child: Icon(Icons.person)),
        title: Text('Ученик: ${link.studentId.substring(0, 8)}…'),
        subtitle: Text('ID: ${link.studentId}'),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            IconButton(
              icon: const Icon(Icons.link_off, color: Colors.red),
              tooltip: 'Отвязать',
              onPressed: () => _confirmUnlink(context, ref),
            ),
            const Icon(Icons.expand_more),
          ],
        ),
        children: [
          _ChildGroupsSection(studentId: link.studentId),
        ],
      ),
    );
  }

  void _confirmUnlink(BuildContext context, WidgetRef ref) {
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Отвязать ребёнка?'),
        content: const Text('Вы больше не будете видеть данные этого ученика.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Отмена'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Colors.red,
            ),
            onPressed: () {
              ref.read(childrenProvider.notifier).unlinkChild(link.studentId);
              Navigator.pop(ctx);
            },
            child: const Text('Отвязать'),
          ),
        ],
      ),
    );
  }
}

// ─── Child groups section ─────────────────────────────────────────────────────

class _ChildGroupsSection extends ConsumerWidget {
  const _ChildGroupsSection({required this.studentId});

  final String studentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final groupsAsync = ref.watch(childGroupsProvider(studentId));

    return groupsAsync.when(
      loading: () => const Padding(
        padding: EdgeInsets.all(16),
        child: Center(child: CircularProgressIndicator()),
      ),
      error: (e, _) => Padding(
        padding: const EdgeInsets.all(16),
        child: Text('Ошибка загрузки групп: $e'),
      ),
      data: (groups) {
        if (groups.isEmpty) {
          return const Padding(
            padding: EdgeInsets.all(16),
            child: Text('Нет групп'),
          );
        }
        return Column(
          children: groups
              .map(
                (g) => ListTile(
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 24, vertical: 0),
                  title: Text(g.name),
                  subtitle: Text(g.subject ?? ''),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      TextButton.icon(
                        icon: const Icon(Icons.grade, size: 18),
                        label: const Text('Оценки'),
                        onPressed: () => Navigator.push<void>(
                          context,
                          MaterialPageRoute(
                            builder: (_) => GradesPage(
                              groupId: g.id,
                              studentId: studentId,
                            ),
                          ),
                        ),
                      ),
                      TextButton.icon(
                        icon: const Icon(Icons.calendar_month, size: 18),
                        label: const Text('Посещ.'),
                        onPressed: () => Navigator.push<void>(
                          context,
                          MaterialPageRoute(
                            builder: (_) => AttendancePage(
                              groupId: g.id,
                              studentId: studentId,
                            ),
                          ),
                        ),
                      ),
                      TextButton.icon(
                        icon: const Icon(Icons.bar_chart, size: 18),
                        label: const Text('ROI'),
                        onPressed: () => Navigator.push<void>(
                          context,
                          MaterialPageRoute(
                            builder: (_) => ParentRoiPage(
                              studentId: studentId,
                              groupId: g.id,
                              studentName: g.name,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              )
              .toList(),
        );
      },
    );
  }
}
