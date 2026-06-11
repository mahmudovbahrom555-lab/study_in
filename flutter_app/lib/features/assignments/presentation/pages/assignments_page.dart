import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/assignment.dart';
import '../providers/assignments_provider.dart';

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
    final isTeacher =
        ref.watch(authProvider).user?.role == 'teacher';

    return Scaffold(
      appBar: AppBar(title: const Text('Задания')),
      floatingActionButton: isTeacher
          ? FloatingActionButton(
              onPressed: () => _showCreateSheet(context),
              child: const Icon(Icons.add),
            )
          : null,
      body: Builder(builder: (_) {
        if (state.isLoading && state.assignments.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }
        if (state.assignments.isEmpty) {
          return const Center(child: Text('Заданий пока нет'));
        }
        return RefreshIndicator(
          onRefresh: () => ref
              .read(assignmentsProvider(widget.groupId).notifier)
              .load(),
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
      }),
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
                ? Icons.check_circle
                : Icons.assignment_outlined,
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
                'Срок: ${fmt.format(assignment.dueDate!)}',
                style: TextStyle(
                  color: overdue ? Colors.red : null,
                  fontSize: 12,
                ),
              ),
            if (assignment.isSubmitted && assignment.submissionGrade != null)
              Text(
                'Оценка: ${assignment.submissionGrade}',
                style: const TextStyle(
                    color: Colors.green, fontWeight: FontWeight.bold),
              ),
          ],
        ),
        isThreeLine: true,
        trailing: isTeacher
            ? IconButton(
                icon: const Icon(Icons.delete_outline, color: Colors.red),
                onPressed: () => ref
                    .read(assignmentsProvider(groupId).notifier)
                    .delete(assignment.id),
              )
            : !assignment.isSubmitted
                ? TextButton(
                    onPressed: () => _showSubmitDialog(context, ref),
                    child: const Text('Сдать'),
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
        title: Text('Сдать: ${assignment.title}'),
        content: TextField(
          controller: ctrl,
          decoration: const InputDecoration(labelText: 'Комментарий (необяз.)'),
          maxLines: 3,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Отмена'),
          ),
          FilledButton(
            onPressed: () {
              ref.read(assignmentsProvider(groupId).notifier).submit(
                    assignment.id,
                    comment: ctrl.text.trim().isEmpty ? null : ctrl.text.trim(),
                  );
              Navigator.pop(ctx);
            },
            child: const Text('Отправить'),
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
          Text('Новое задание',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 16),
          TextField(
            controller: _titleCtrl,
            decoration: const InputDecoration(
              labelText: 'Название *',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _descCtrl,
            decoration: const InputDecoration(
              labelText: 'Описание',
              border: OutlineInputBorder(),
            ),
            maxLines: 3,
          ),
          const SizedBox(height: 12),
          OutlinedButton.icon(
            icon: const Icon(Icons.calendar_today),
            label: Text(_dueDate == null
                ? 'Срок сдачи (необязательно)'
                : fmt.format(_dueDate!)),
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
                description:
                    _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
                dueDate: _dueDate,
              );
              Navigator.pop(context);
            },
            child: const Text('Создать'),
          ),
        ],
      ),
    );
  }
}
