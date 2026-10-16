import 'package:flutter/material.dart';

import '../../../../core/localization/l10n.dart';
import '../../../groups/domain/entities/group.dart';

/// Сохраняет оценку; возвращает текст ошибки или null при успехе.
typedef SubmitGrade = Future<String?> Function({
  required String studentId,
  required String subject,
  required double value,
  required double maxValue,
  String? comment,
});

/// «4,5» и «4.5» — одно и то же: на русской клавиатуре десятичная запятая.
double? parseScore(String? text) =>
    double.tryParse((text ?? '').trim().replaceAll(',', '.'));

/// Форма оценки: ученик из списка группы, «за что», балл и максимум.
/// Закрывается только после успешного сохранения.
class AddGradeSheet extends StatefulWidget {
  const AddGradeSheet({super.key, required this.students, required this.onSubmit});

  final List<GroupMember> students;
  final SubmitGrade onSubmit;

  @override
  State<AddGradeSheet> createState() => _AddGradeSheetState();
}

class _AddGradeSheetState extends State<AddGradeSheet> {
  // «За что» и максимум запоминаются до перезапуска: репетитор обычно
  // оценивает подряд всю группу за одну работу.
  static String _lastSubject = '';
  static String _lastMax = '100';

  final _formKey = GlobalKey<FormState>();
  final _subjectCtrl = TextEditingController(text: _lastSubject);
  final _valueCtrl = TextEditingController();
  final _maxValueCtrl = TextEditingController(text: _lastMax);
  final _commentCtrl = TextEditingController();
  String? _studentId;
  bool _saving = false;
  String? _error;

  @override
  void dispose() {
    _subjectCtrl.dispose();
    _valueCtrl.dispose();
    _maxValueCtrl.dispose();
    _commentCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_saving || !_formKey.currentState!.validate()) return;
    setState(() {
      _saving = true;
      _error = null;
    });
    final comment = _commentCtrl.text.trim();
    final error = await widget.onSubmit(
      studentId: _studentId!,
      subject: _subjectCtrl.text.trim(),
      value: parseScore(_valueCtrl.text)!,
      maxValue: parseScore(_maxValueCtrl.text)!,
      comment: comment.isEmpty ? null : comment,
    );
    if (!mounted) return;
    if (error != null) {
      setState(() {
        _saving = false;
        _error = error;
      });
      return;
    }
    _lastSubject = _subjectCtrl.text.trim();
    _lastMax = _maxValueCtrl.text.trim();
    Navigator.of(context).pop();
  }

  String? _validateValue(String? v) {
    final l10n = context.l10n;
    final n = parseScore(v);
    if (n == null) return l10n.mustBeNumber;
    if (n < 0) return '≥ 0';
    final max = parseScore(_maxValueCtrl.text);
    if (max != null && n > max) return l10n.gradeAboveMax(_maxValueCtrl.text.trim());
    return null;
  }

  String? _validateMax(String? v) {
    final n = parseScore(v);
    if (n == null) return context.l10n.mustBeNumber;
    if (n <= 0) return '> 0';
    return null;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final insets = MediaQuery.viewInsetsOf(context);
    final theme = Theme.of(context);

    return Padding(
      padding: EdgeInsets.fromLTRB(16, 16, 16, 16 + insets.bottom),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(l10n.addGrade, style: theme.textTheme.titleLarge),
            const SizedBox(height: 16),
            DropdownButtonFormField<String>(
              initialValue: _studentId,
              isExpanded: true,
              decoration: InputDecoration(labelText: l10n.studentRequired),
              items: [
                for (final s in widget.students)
                  DropdownMenuItem(value: s.studentId, child: Text(s.name)),
              ],
              onChanged: _saving ? null : (v) => setState(() => _studentId = v),
              validator: (v) => v == null ? l10n.requiredField : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _subjectCtrl,
              enabled: !_saving,
              maxLength: 100,
              decoration: InputDecoration(
                labelText: l10n.gradeForRequired,
                hintText: l10n.gradeForHint,
                counterText: '',
              ),
              validator: (v) =>
                  v == null || v.trim().isEmpty ? l10n.requiredField : null,
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: 12),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _valueCtrl,
                    enabled: !_saving,
                    decoration: InputDecoration(labelText: l10n.gradeRequired),
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    validator: _validateValue,
                    textInputAction: TextInputAction.next,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _maxValueCtrl,
                    enabled: !_saving,
                    decoration: InputDecoration(labelText: l10n.maxShort),
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    validator: _validateMax,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _commentCtrl,
              enabled: !_saving,
              decoration: InputDecoration(labelText: l10n.commentOptional),
              maxLines: 2,
            ),
            if (_error != null) ...[
              const SizedBox(height: 12),
              Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
            ],
            const SizedBox(height: 20),
            FilledButton(
              onPressed: _saving ? null : _submit,
              child: _saving
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(l10n.save),
            ),
          ],
        ),
      ),
    );
  }
}
