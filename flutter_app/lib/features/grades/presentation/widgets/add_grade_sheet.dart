import 'package:flutter/material.dart';

class AddGradeSheet extends StatefulWidget {
  const AddGradeSheet({super.key, required this.onSubmit});

  final void Function({
    required String studentId,
    required String subject,
    required double value,
    double maxValue,
    String? comment,
  }) onSubmit;

  @override
  State<AddGradeSheet> createState() => _AddGradeSheetState();
}

class _AddGradeSheetState extends State<AddGradeSheet> {
  final _formKey = GlobalKey<FormState>();
  final _studentCtrl = TextEditingController();
  final _subjectCtrl = TextEditingController();
  final _valueCtrl = TextEditingController();
  final _maxValueCtrl = TextEditingController(text: '100');
  final _commentCtrl = TextEditingController();

  @override
  void dispose() {
    _studentCtrl.dispose();
    _subjectCtrl.dispose();
    _valueCtrl.dispose();
    _maxValueCtrl.dispose();
    _commentCtrl.dispose();
    super.dispose();
  }

  void _submit() {
    if (!_formKey.currentState!.validate()) return;
    widget.onSubmit(
      studentId: _studentCtrl.text.trim(),
      subject: _subjectCtrl.text.trim(),
      value: double.parse(_valueCtrl.text.trim()),
      maxValue: double.tryParse(_maxValueCtrl.text.trim()) ?? 100,
      comment: _commentCtrl.text.trim().isEmpty
          ? null
          : _commentCtrl.text.trim(),
    );
    Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final insets = MediaQuery.viewInsetsOf(context);
    return Padding(
      padding: EdgeInsets.fromLTRB(16, 16, 16, 16 + insets.bottom),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('Добавить оценку',
                style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 16),
            TextFormField(
              controller: _studentCtrl,
              decoration: const InputDecoration(labelText: 'ID студента *'),
              validator: (v) =>
                  v == null || v.trim().isEmpty ? 'Обязательное поле' : null,
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _subjectCtrl,
              decoration: const InputDecoration(labelText: 'Предмет *'),
              validator: (v) =>
                  v == null || v.trim().isEmpty ? 'Обязательное поле' : null,
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _valueCtrl,
                    decoration: const InputDecoration(labelText: 'Оценка *'),
                    keyboardType: const TextInputType.numberWithOptions(
                        decimal: true),
                    validator: (v) {
                      final n = double.tryParse(v ?? '');
                      if (n == null) return 'Число';
                      if (n < 0) return '≥ 0';
                      return null;
                    },
                    textInputAction: TextInputAction.next,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _maxValueCtrl,
                    decoration: const InputDecoration(labelText: 'Макс.'),
                    keyboardType: const TextInputType.numberWithOptions(
                        decimal: true),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _commentCtrl,
              decoration:
                  const InputDecoration(labelText: 'Комментарий (необяз.)'),
              maxLines: 2,
            ),
            const SizedBox(height: 20),
            FilledButton(onPressed: _submit, child: const Text('Сохранить')),
          ],
        ),
      ),
    );
  }
}
