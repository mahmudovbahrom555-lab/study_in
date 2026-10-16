import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/theme/colors.dart';
import '../../../../core/widgets/error_view.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../groups/domain/entities/group.dart';
import '../../../groups/presentation/providers/groups_provider.dart';
import '../../domain/entities/grade.dart';
import '../providers/grades_provider.dart';
import '../widgets/add_grade_sheet.dart';

class GradesPage extends ConsumerStatefulWidget {
  const GradesPage({
    super.key,
    required this.groupId,
    this.isTeacher = false,
  });

  final String groupId;
  final bool isTeacher;

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
    final params = (
      groupId: widget.groupId,
      studentId: ref.watch(authProvider).user?.id ?? '',
    );
    return GradeHistoryView(
      title: context.l10n.myGrades,
      grades: ref.watch(studentGradesProvider(params)),
      onRetry: () => ref.invalidate(studentGradesProvider(params)),
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
    final notifier = ref.read(groupGradesProvider(groupId).notifier);

    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.gradebook)),
      floatingActionButton: FloatingActionButton(
        tooltip: context.l10n.addGrade,
        onPressed: () => _showAdd(context, ref),
        child: const Icon(LucideIcons.plus),
      ),
      body: Builder(
        builder: (_) {
          if (state.isLoading && state.grades.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state.error != null && state.grades.isEmpty) {
            return ErrorView(message: state.error, onRetry: notifier.load);
          }
          if (state.grades.isEmpty) {
            return Center(child: Text(context.l10n.noGrades));
          }
          return RefreshIndicator(
            onRefresh: notifier.load,
            child: ListView.builder(
              padding: const EdgeInsets.only(bottom: 88), // не под кнопкой «+»
              itemCount: state.grades.length,
              itemBuilder: (_, i) {
                final g = state.grades[i];
                return _GradeTile(
                  grade: g,
                  showStudent: true,
                  // Оценку за ДЗ меняют в проверке работы, не в журнале.
                  onDelete: g.isHomework ? null : () => _confirmDelete(context, ref, g),
                );
              },
            ),
          );
        },
      ),
    );
  }

  Future<void> _showAdd(BuildContext context, WidgetRef ref) async {
    // Ученики для выбора — из списка группы (обычно уже загружен на её странице).
    if (ref.read(membersProvider(groupId)).valueOrNull == null) {
      await ref.read(membersProvider(groupId).notifier).load();
    }
    if (!context.mounted) return;
    final membersState = ref.read(membersProvider(groupId));
    final members = membersState.valueOrNull ?? const <GroupMember>[];
    if (members.isEmpty) {
      final text = membersState.hasError
          ? userErrorMessage(membersState.error!)
          : context.l10n.noStudentsToGrade;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(text)));
      return;
    }
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => AddGradeSheet(
        students: members,
        onSubmit: ref.read(groupGradesProvider(groupId).notifier).createGrade,
      ),
    );
  }

  Future<void> _confirmDelete(BuildContext context, WidgetRef ref, Grade g) async {
    final l10n = context.l10n;
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.deleteGradeQ),
        content: Text(
          l10n.deleteGradeBody(
            g.studentName.isEmpty ? '—' : g.studentName,
            '${formatScore(g.value, l10n.localeName)} / ${formatScore(g.maxValue, l10n.localeName)}',
            g.subject,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(l10n.cancel),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(
              l10n.delete,
              style: TextStyle(color: Theme.of(ctx).colorScheme.error),
            ),
          ),
        ],
      ),
    );
    if (ok != true) return;
    final error = await ref.read(groupGradesProvider(groupId).notifier).deleteGrade(g.id);
    if (error != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error)));
    }
  }
}

/// Балл без лишних нулей и с десятичным разделителем языка: «4,5», «85».
String formatScore(double v, String locale) =>
    NumberFormat.decimalPattern(locale).format(double.parse(v.toStringAsFixed(2)));

// ─── Grade history (student, parent) ─────────────────────────────────────────

/// Оценки одного ученика. Данные приходят снаружи: ученик берёт свои,
/// родитель — ребёнка через свой модуль (у них разные эндпоинты и права).
class GradeHistoryView extends StatelessWidget {
  const GradeHistoryView({
    super.key,
    required this.title,
    required this.grades,
    required this.onRetry,
  });

  final String title;
  final AsyncValue<List<Grade>> grades;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: grades.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => ErrorView(error: e, onRetry: onRetry),
        data: (grades) {
          if (grades.isEmpty) {
            return Center(child: Text(context.l10n.noGrades));
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
    final avg =
        grades.map((g) => g.percentage).reduce((a, b) => a + b) / grades.length;
    return Container(
      color: Theme.of(context).colorScheme.primaryContainer,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        children: [
          const Icon(LucideIcons.chartColumn),
          const SizedBox(width: 8),
          Text(
            context.l10n.gradesSummary(
              formatScore(double.parse(avg.toStringAsFixed(1)), context.l10n.localeName),
              grades.length,
            ),
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
    final l10n = context.l10n;
    final theme = Theme.of(context);
    final locale = l10n.localeName;
    final pct = grade.percentage;
    final color = pct >= 80
        ? AppColors.success
        : pct >= 60
            ? AppColors.warning
            : AppColors.error;
    final score = '${formatScore(grade.value, locale)} / ${formatScore(grade.maxValue, locale)}';
    final date = DateFormat('dd.MM.yyyy').format(grade.gradedAt.toLocal());

    // В журнале репетитора строка — ученик, у ученика и родителя — за что оценка.
    final title = showStudent
        ? (grade.studentName.isEmpty ? '—' : grade.studentName)
        : grade.subject;
    final details = showStudent ? '${grade.subject} · $score · $date' : '$score · $date';

    return ListTile(
      leading: CircleAvatar(
        backgroundColor: color.withValues(alpha: 0.15),
        child: Text(
          formatScore(grade.value, locale),
          style: TextStyle(color: color, fontWeight: FontWeight.bold),
        ),
      ),
      title: Row(
        children: [
          Flexible(child: Text(title, overflow: TextOverflow.ellipsis)),
          if (grade.isHomework) ...[
            const SizedBox(width: 8),
            _KindChip(label: l10n.homeworkShort),
          ],
        ],
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(details),
          if (grade.comment != null)
            Text(grade.comment!, style: theme.textTheme.bodySmall),
        ],
      ),
      isThreeLine: grade.comment != null,
      trailing: onDelete != null
          ? IconButton(
              tooltip: l10n.deleteGrade,
              icon: Icon(LucideIcons.trash2, color: theme.colorScheme.error),
              onPressed: onDelete,
            )
          : null,
    );
  }
}

class _KindChip extends StatelessWidget {
  const _KindChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
      decoration: BoxDecoration(
        color: scheme.secondaryContainer,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(color: scheme.onSecondaryContainer),
      ),
    );
  }
}
