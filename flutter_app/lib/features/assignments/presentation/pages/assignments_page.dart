import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/router/routes.dart';
import '../../../../core/theme/colors.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/assignment.dart';
import '../providers/assignments_provider.dart';
import '../../../../core/localization/l10n.dart';

class AssignmentsPage extends ConsumerStatefulWidget {
  const AssignmentsPage({super.key, required this.groupId});

  final String groupId;

  @override
  ConsumerState<AssignmentsPage> createState() => _AssignmentsPageState();
}

class _AssignmentsPageState extends ConsumerState<AssignmentsPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(assignmentsProvider(widget.groupId).notifier).load(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(assignmentsProvider(widget.groupId));
    final isTeacher = ref.watch(authProvider).user?.role == 'teacher';

    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.tabAssignments)),
      floatingActionButton: isTeacher
          ? FloatingActionButton(
            tooltip: context.l10n.createAssignment,
              onPressed: () => _showCreateSheet(context),
              child: const Icon(LucideIcons.plus),
            )
          : null,
      body: Builder(
        builder: (_) {
          if (state.isLoading && state.assignments.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state.assignments.isEmpty) {
            return Center(child: Text(context.l10n.noAssignments));
          }
          return RefreshIndicator(
            onRefresh: () =>
                ref.read(assignmentsProvider(widget.groupId).notifier).load(),
            child: ListView.builder(
              padding: const EdgeInsets.all(12),
              itemCount: state.assignments.length,
              itemBuilder: (_, i) => _AssignmentCard(
                assignment: state.assignments[i],
                isTeacher: isTeacher,
                groupId: widget.groupId,
              ),
            ),
          );
        },
      ),
    );
  }

  void _showCreateSheet(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _CreateAssignmentSheet(
        onSubmit: ({required title, description, dueDate}) {
          ref.read(assignmentsProvider(widget.groupId).notifier).create(
                title: title,
                description: description,
                dueDate: dueDate,
              );
        },
      ),
    );
  }
}

// ─── Assignment card ──────────────────────────────────────────────────────────

class _AssignmentCard extends ConsumerWidget {
  const _AssignmentCard({
    required this.assignment,
    required this.isTeacher,
    required this.groupId,
  });

  final Assignment assignment;
  final bool isTeacher;
  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    final a = assignment;
    final (statusIcon, statusColor, statusText) = _status(context);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => context.push(Routes.assignment(groupId, a.id)),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 4, 12),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.only(top: 2),
                child: Icon(statusIcon, color: statusColor),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(a.title, style: theme.textTheme.titleMedium),
                    if (a.description != null)
                      Text(
                        a.description!,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.bodyMedium,
                      ),
                    if (a.dueDate != null)
                      Text(
                        l10n.dueDate(DateFormat('dd.MM.yyyy HH:mm').format(a.dueDate!)),
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: a.isOverdue ? theme.colorScheme.error : null,
                        ),
                      ),
                    const SizedBox(height: 4),
                    Text(
                      statusText,
                      style: theme.textTheme.labelLarge?.copyWith(color: statusColor),
                    ),
                  ],
                ),
              ),
              if (isTeacher)
                PopupMenuButton<void>(
                  icon: const Icon(LucideIcons.ellipsisVertical),
                  itemBuilder: (_) => [
                    PopupMenuItem(
                      onTap: () => ref
                          .read(assignmentsProvider(groupId).notifier)
                          .delete(a.id),
                      child: Text(
                        l10n.deleteAssignment,
                        style: TextStyle(color: theme.colorScheme.error),
                      ),
                    ),
                  ],
                ),
            ],
          ),
        ),
      ),
    );
  }

  /// Статус всегда текстом, цвет — только подсказка.
  (IconData, Color, String) _status(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final l10n = context.l10n;
    final a = assignment;
    if (isTeacher) {
      final st = a.stats;
      final toReview = st?.ungraded ?? 0;
      final text = [
        l10n.hwSubmittedCount(st?.submitted ?? 0),
        if (toReview > 0) l10n.hwToReview(toReview),
      ].join(' · ');
      return (
        LucideIcons.fileCheck,
        toReview > 0 ? scheme.primary : scheme.onSurfaceVariant,
        text,
      );
    }
    final mine = a.mySubmission;
    if (mine?.grade != null) {
      return (LucideIcons.circleCheck, AppColors.success, l10n.gradeValue('${mine!.grade}'));
    }
    if (mine != null) return (LucideIcons.clock, scheme.primary, l10n.hwSubmitted);
    if (a.isOverdue) return (LucideIcons.circleAlert, scheme.error, l10n.hwOverdue);
    return (LucideIcons.fileCheck, scheme.onSurfaceVariant, l10n.hwNotSubmitted);
  }
}

// ─── Create assignment sheet ──────────────────────────────────────────────────

class _CreateAssignmentSheet extends StatefulWidget {
  const _CreateAssignmentSheet({required this.onSubmit});

  final void Function({
    required String title,
    String? description,
    DateTime? dueDate,
  }) onSubmit;

  @override
  State<_CreateAssignmentSheet> createState() => _CreateAssignmentSheetState();
}

class _CreateAssignmentSheetState extends State<_CreateAssignmentSheet> {
  final _titleCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  DateTime? _dueDate;

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final fmt = DateFormat('dd.MM.yyyy HH:mm');
    return Padding(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 16,
        bottom: MediaQuery.viewInsetsOf(context).bottom + 16,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            context.l10n.newAssignment,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _titleCtrl,
            decoration: InputDecoration(
              labelText: context.l10n.titleRequired,
              border: const OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _descCtrl,
            decoration: InputDecoration(
              labelText: context.l10n.description,
              border: const OutlineInputBorder(),
            ),
            maxLines: 3,
          ),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            icon: const Icon(LucideIcons.calendar),
            label: Text(
              _dueDate == null
                  ? context.l10n.dueDateOptional
                  : fmt.format(_dueDate!),
            ),
            onPressed: () async {
              final date = await showDatePicker(
                context: context,
                initialDate: DateTime.now().add(const Duration(days: 7)),
                firstDate: DateTime.now(),
                lastDate: DateTime.now().add(const Duration(days: 365)),
              );
              // Срок — конец выбранного дня, а не полночь в его начале.
              if (date != null) {
                setState(() => _dueDate = DateTime(date.year, date.month, date.day, 23, 59));
              }
            },
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: () {
              final title = _titleCtrl.text.trim();
              if (title.isEmpty) return;
              widget.onSubmit(
                title: title,
                description: _descCtrl.text.trim().isEmpty
                    ? null
                    : _descCtrl.text.trim(),
                dueDate: _dueDate,
              );
              Navigator.pop(context);
            },
            child: Text(context.l10n.create),
          ),
        ],
      ),
    );
  }
}
