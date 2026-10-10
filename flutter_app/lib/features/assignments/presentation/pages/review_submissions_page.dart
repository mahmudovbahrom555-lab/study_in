import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/widgets/error_view.dart';
import '../../domain/entities/assignment.dart';
import '../../domain/entities/submission.dart';
import '../providers/assignments_provider.dart';
import '../widgets/attachment_views.dart';

/// Работы учеников по заданию: сначала непроверенные. Репетитор смотрит
/// файлы и ставит оценку прямо в карточке.
class ReviewSubmissionsPage extends ConsumerStatefulWidget {
  const ReviewSubmissionsPage({super.key, required this.assignment});

  final Assignment assignment;

  @override
  ConsumerState<ReviewSubmissionsPage> createState() =>
      _ReviewSubmissionsPageState();
}

class _ReviewSubmissionsPageState extends ConsumerState<ReviewSubmissionsPage> {
  late Future<List<Submission>> _future = _fetch();

  Future<List<Submission>> _fetch() => ref
      .read(assignmentsRepositoryProvider)
      .listSubmissions(widget.assignment.groupId, widget.assignment.id);

  void _reload() {
    setState(() => _future = _fetch());
    ref.read(assignmentsProvider(widget.assignment.groupId).notifier).load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      appBar: AppBar(title: Text(widget.assignment.title)),
      body: FutureBuilder<List<Submission>>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snap.hasError) {
            return ErrorView(error: snap.error, onRetry: _reload);
          }
          final subs = snap.data!;
          if (subs.isEmpty) return Center(child: Text(l10n.hwNoSubmissions));
          return RefreshIndicator(
            onRefresh: () async => _reload(),
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: subs.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (_, i) => _SubmissionCard(
                assignment: widget.assignment,
                submission: subs[i],
                onGraded: _reload,
              ),
            ),
          );
        },
      ),
    );
  }
}

class _SubmissionCard extends StatelessWidget {
  const _SubmissionCard({
    required this.assignment,
    required this.submission,
    required this.onGraded,
  });

  final Assignment assignment;
  final Submission submission;
  final VoidCallback onGraded;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    final s = submission;
    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(s.studentName ?? '—', style: theme.textTheme.titleMedium),
            Text(
              l10n.hwSubmittedAt(DateFormat('dd.MM.yyyy HH:mm').format(s.submittedAt)),
              style: theme.textTheme.bodySmall,
            ),
            if (s.comment case final c? when c.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(c),
            ],
            if (s.files.isNotEmpty) ...[
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [for (final f in s.files) AttachmentChip(file: f)],
              ),
            ],
            const Divider(height: 24),
            _GradeForm(
              assignment: assignment,
              submission: s,
              onGraded: onGraded,
            ),
          ],
        ),
      ),
    );
  }
}

/// Оценка: у оценённой работы — значение и кнопка «Изменить», иначе форма сразу.
class _GradeForm extends ConsumerStatefulWidget {
  const _GradeForm({
    required this.assignment,
    required this.submission,
    required this.onGraded,
  });

  final Assignment assignment;
  final Submission submission;
  final VoidCallback onGraded;

  @override
  ConsumerState<_GradeForm> createState() => _GradeFormState();
}

class _GradeFormState extends ConsumerState<_GradeForm> {
  late final _grade =
      TextEditingController(text: widget.submission.grade?.toString());
  late final _note = TextEditingController(text: widget.submission.teacherNote);
  late bool _editing = !widget.submission.isGraded;
  bool _saving = false;
  String? _gradeError;

  @override
  void dispose() {
    _grade.dispose();
    _note.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final l10n = context.l10n;
    final value = int.tryParse(_grade.text.trim());
    if (value == null || value < 0 || value > 100) {
      setState(() => _gradeError = l10n.hwGradeInvalid);
      return;
    }
    final note = _note.text.trim();
    setState(() {
      _gradeError = null;
      _saving = true;
    });
    try {
      await ref.read(assignmentsRepositoryProvider).grade(
            widget.assignment.groupId,
            widget.assignment.id,
            widget.submission.id,
            grade: value,
            note: note.isEmpty ? null : note,
          );
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(l10n.hwGradeSaved)));
      widget.onGraded();
    } catch (e) {
      if (!mounted) return;
      setState(() => _saving = false);
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    final s = widget.submission;

    if (!_editing) {
      return Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.gradeValue('${s.grade}'),
                  style: theme.textTheme.titleMedium?.copyWith(color: theme.colorScheme.primary),
                ),
                if (s.teacherNote case final n? when n.isNotEmpty) Text(n),
              ],
            ),
          ),
          TextButton(
            onPressed: () => setState(() => _editing = true),
            child: Text(l10n.hwChangeGrade),
          ),
        ],
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _grade,
          keyboardType: TextInputType.number,
          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          maxLength: 3,
          decoration: InputDecoration(
            labelText: l10n.hwGradeLabel,
            errorText: _gradeError,
            counterText: '',
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _note,
          minLines: 1,
          maxLines: 3,
          decoration: InputDecoration(labelText: l10n.hwNoteOptional),
        ),
        const SizedBox(height: 12),
        FilledButton(
          onPressed: _saving ? null : _save,
          child: _saving
              ? const SizedBox.square(
                  dimension: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : Text(l10n.hwGradeAction),
        ),
      ],
    );
  }
}
