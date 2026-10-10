import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

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
    final fmt = DateFormat('dd.MM.yyyy HH:mm');
    final overdue = assignment.dueDate != null &&
        assignment.dueDate!.isBefore(DateTime.now()) &&
        !assignment.isSubmitted;

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: assignment.isSubmitted
              ? Colors.green.withOpacity(0.15)
              : overdue
                  ? Colors.red.withOpacity(0.15)
                  : null,
          child: Icon(
            assignment.isSubmitted
                ? LucideIcons.circleCheck
                : LucideIcons.fileCheck,
            color: assignment.isSubmitted
                ? Colors.green
                : overdue
                    ? Colors.red
                    : null,
          ),
        ),
        title: Text(assignment.title),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (assignment.description != null)
              Text(
                assignment.description!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            if (assignment.dueDate != null)
              Text(
                context.l10n.dueDate(fmt.format(assignment.dueDate!)),
                style: TextStyle(
                  color: overdue ? Colors.red : null,
                  fontSize: 12,
                ),
              ),
            if (assignment.isSubmitted && assignment.submissionGrade != null)
              Text(
                context.l10n.gradeValue('${assignment.submissionGrade}'),
                style: const TextStyle(
                  color: Colors.green,
                  fontWeight: FontWeight.bold,
                ),
              ),
          ],
        ),
        isThreeLine: true,
        trailing: isTeacher
            ? IconButton(
              tooltip: context.l10n.deleteAssignment,
                icon: const Icon(LucideIcons.trash2, color: Colors.red),
                onPressed: () => ref
                    .read(assignmentsProvider(groupId).notifier)
                    .delete(assignment.id),
              )
            : !assignment.isSubmitted
                ? TextButton(
                    onPressed: () => _showSubmitDialog(context, ref),
                    child: Text(context.l10n.submit),
                  )
                : null,
      ),
    );
  }

  void _showSubmitDialog(BuildContext context, WidgetRef ref) {
    final ctrl = TextEditingController();
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(context.l10n.submitTitle(assignment.title)),
        content: TextField(
          controller: ctrl,
          decoration: InputDecoration(labelText: context.l10n.commentOptional),
          maxLines: 3,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: Text(context.l10n.cancel),
          ),
          FilledButton(
            onPressed: () {
              ref.read(assignmentsProvider(groupId).notifier).submit(
                    assignment.id,
                    comment: ctrl.text.trim().isEmpty ? null : ctrl.text.trim(),
                  );
              Navigator.pop(ctx);
            },
            child: Text(context.l10n.send),
          ),
        ],
      ),
    );
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
              if (date != null) setState(() => _dueDate = date);
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
