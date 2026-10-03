import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../attendance/presentation/pages/attendance_page.dart';
import '../../../grades/presentation/pages/grades_page.dart';
import '../../../reports/presentation/pages/parent_roi_page.dart';
import '../../domain/entities/parent_link.dart';
import '../providers/parents_provider.dart';
import '../../../../core/widgets/error_view.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/localization/l10n.dart';

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
      appBar: AppBar(title: Text(context.l10n.myChildren)),
      floatingActionButton: FloatingActionButton.extended(
        icon: const Icon(Icons.link),
        label: Text(context.l10n.add),
        onPressed: () => _showLinkDialog(context),
      ),
      body: Builder(
        builder: (_) {
          if (state.isLoading && state.children.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state.error != null && state.children.isEmpty) {
            return ErrorView(
              message: state.error!,
              onRetry: () => ref.read(childrenProvider.notifier).load(),
            );
          }
          if (state.children.isEmpty) {
            return Center(
              child: Text(context.l10n.noChildren),
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
        title: Text(context.l10n.linkChild),
        content: TextField(
          controller: ctrl,
          decoration: InputDecoration(
            labelText: context.l10n.studentId,
            hintText: context.l10n.studentUuidHint,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(context.l10n.cancel),
          ),
          FilledButton(
            onPressed: () {
              final id = ctrl.text.trim();
              if (id.isNotEmpty) {
                ref.read(childrenProvider.notifier).linkChild(id);
              }
              Navigator.pop(ctx);
            },
            child: Text(context.l10n.linkAction),
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
        title: Text(context.l10n.studentShort(link.studentId.substring(0, 8))),
        subtitle: Text('ID: ${link.studentId}'),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            IconButton(
              icon: const Icon(Icons.link_off, color: Colors.red),
              tooltip: context.l10n.unlink,
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
        title: Text(context.l10n.unlinkChildQ),
        content: Text(context.l10n.unlinkChildBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(context.l10n.cancel),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: Colors.red,
            ),
            onPressed: () {
              ref.read(childrenProvider.notifier).unlinkChild(link.studentId);
              Navigator.pop(ctx);
            },
            child: Text(context.l10n.unlink),
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
        child: Text(context.l10n.groupsLoadFailed(userErrorMessage(e))),
      ),
      data: (groups) {
        if (groups.isEmpty) {
          return Padding(
            padding: const EdgeInsets.all(16),
            child: Text(context.l10n.noGroups),
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
                        label: Text(context.l10n.tabGrades),
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
                        label: Text(context.l10n.tabAttendanceShort),
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
