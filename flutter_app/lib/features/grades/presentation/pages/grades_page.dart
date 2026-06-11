import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../domain/entities/grade.dart';
import '../providers/grades_provider.dart';
import '../widgets/add_grade_sheet.dart';

class GradesPage extends ConsumerStatefulWidget {
  const GradesPage({
    super.key,
    required this.groupId,
    this.isTeacher = false,
    this.studentId,
  });

  final String groupId;
  final bool isTeacher;
  final String? studentId;

  @override
  ConsumerState<GradesPage> createState() => _GradesPageState();
}

class _GradesPageState extends ConsumerState<GradesPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(_load);
  }

  void _load() {
    if (widget.isTeacher) {
      ref.read(groupGradesProvider(widget.groupId).notifier).load();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (widget.isTeacher) {
      return _TeacherView(groupId: widget.groupId);
    }
    return _StudentView(
      groupId: widget.groupId,
      studentId: widget.studentId ?? '',
    );
  }
}

// ─── Teacher view ─────────────────────────────────────────────────────────────

class _TeacherView extends ConsumerWidget {
  const _TeacherView({required this.groupId});

  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(groupGradesProvider(groupId));

    return Scaffold(
      appBar: AppBar(title: const Text('Журнал оценок')),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showAdd(context, ref),
        child: const Icon(Icons.add),
      ),
      body: Builder(builder: (_) {
        if (state.isLoading && state.grades.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }
        if (state.grades.isEmpty) {
          return const Center(child: Text('Оценок пока нет'));
        }
        return RefreshIndicator(
          onRefresh: () =>
              ref.read(groupGradesProvider(groupId).notifier).load(),
          child: ListView.builder(
            itemCount: state.grades.length,
            itemBuilder: (_, i) => _GradeTile(
              grade: state.grades[i],
              showStudent: true,
              onDelete: () => ref
                  .read(groupGradesProvider(groupId).notifier)
                  .deleteGrade(state.grades[i].id),
            ),
          ),
        );
      }),
    );
  }

  void _showAdd(BuildContext context, WidgetRef ref) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => AddGradeSheet(
        onSubmit: ({
          required String studentId,
          required String subject,
          required double value,
          double maxValue = 100,
          String? comment,
        }) {
          ref.read(groupGradesProvider(groupId).notifier).createGrade(
                studentId: studentId,
                subject: subject,
                value: value,
                maxValue: maxValue,
                comment: comment,
              );
        },
      ),
    );
  }
}

// ─── Student view ─────────────────────────────────────────────────────────────

class _StudentView extends ConsumerWidget {
  const _StudentView({required this.groupId, required this.studentId});

  final String groupId;
  final String studentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(
      studentGradesProvider((groupId: groupId, studentId: studentId)),
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Мои оценки')),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (grades) {
          if (grades.isEmpty) {
            return const Center(child: Text('Оценок пока нет'));
          }
          return Column(
            children: [
              _SummaryBar(grades: grades),
              Expanded(
                child: ListView.builder(
                  itemCount: grades.length,
                  itemBuilder: (_, i) =>
                      _GradeTile(grade: grades[i], showStudent: false),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

// ─── Summary bar ─────────────────────────────────────────────────────────────

class _SummaryBar extends StatelessWidget {
  const _SummaryBar({required this.grades});

  final List<Grade> grades;

  @override
  Widget build(BuildContext context) {
    if (grades.isEmpty) return const SizedBox.shrink();
    final avg = grades.map((g) => g.percentage).reduce((a, b) => a + b) /
        grades.length;
    return Container(
      color: Theme.of(context).colorScheme.primaryContainer,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        children: [
          const Icon(Icons.bar_chart),
          const SizedBox(width: 8),
          Text(
            'Средний балл: ${avg.toStringAsFixed(1)}%  |  '
            'Всего: ${grades.length}',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      ),
    );
  }
}

// ─── Grade tile ───────────────────────────────────────────────────────────────

class _GradeTile extends StatelessWidget {
  const _GradeTile({
    required this.grade,
    required this.showStudent,
    this.onDelete,
  });

  final Grade grade;
  final bool showStudent;
  final VoidCallback? onDelete;

  @override
  Widget build(BuildContext context) {
    final fmt = DateFormat('dd.MM.yyyy');
    final pct = grade.percentage;
    final color = pct >= 80
        ? Colors.green
        : pct >= 60
            ? Colors.orange
            : Colors.red;

    return ListTile(
      leading: CircleAvatar(
        backgroundColor: color.withOpacity(0.15),
        child: Text(
          grade.value.toStringAsFixed(0),
          style: TextStyle(color: color, fontWeight: FontWeight.bold),
        ),
      ),
      title: Text(grade.subject),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('${grade.value} / ${grade.maxValue}  •  ${fmt.format(grade.gradedAt)}'),
          if (showStudent)
            Text(
              'Студент: ${grade.studentId}',
              style: Theme.of(context).textTheme.labelSmall,
            ),
          if (grade.comment != null)
            Text(
              grade.comment!,
              style: Theme.of(context).textTheme.bodySmall,
            ),
        ],
      ),
      isThreeLine: showStudent || grade.comment != null,
      trailing: onDelete != null
          ? IconButton(
              icon: const Icon(Icons.delete_outline, color: Colors.red),
              onPressed: onDelete,
            )
          : null,
    );
  }
}
