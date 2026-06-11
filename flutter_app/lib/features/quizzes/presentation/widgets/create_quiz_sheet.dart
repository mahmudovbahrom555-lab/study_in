import 'package:flutter/material.dart';

class CreateQuizSheet extends StatefulWidget {
  const CreateQuizSheet({super.key, required this.onSubmit});

  final void Function({
    required String title,
    String? description,
    int maxAttempts,
    int? timeLimit,
  }) onSubmit;

  @override
  State<CreateQuizSheet> createState() => _CreateQuizSheetState();
}

class _CreateQuizSheetState extends State<CreateQuizSheet> {
  final _titleCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  final _attemptsCtrl = TextEditingController(text: '1');
  final _timeLimitCtrl = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descCtrl.dispose();
    _attemptsCtrl.dispose();
    _timeLimitCtrl.dispose();
    super.dispose();
  }

  void _submit() {
    if (!_formKey.currentState!.validate()) return;
    final timeLimit = _timeLimitCtrl.text.trim().isEmpty
        ? null
        : int.tryParse(_timeLimitCtrl.text.trim());
    final attempts = int.tryParse(_attemptsCtrl.text.trim()) ?? 1;
    widget.onSubmit(
      title: _titleCtrl.text.trim(),
      description:
          _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
      maxAttempts: attempts,
      timeLimit: timeLimit,
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
            Text('Новый тест', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 16),
            TextFormField(
              controller: _titleCtrl,
              decoration: const InputDecoration(labelText: 'Название *'),
              validator: (v) =>
                  v == null || v.trim().isEmpty ? 'Обязательное поле' : null,
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _descCtrl,
              decoration:
                  const InputDecoration(labelText: 'Описание (необязательно)'),
              maxLines: 2,
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _attemptsCtrl,
                    decoration:
                        const InputDecoration(labelText: 'Попыток'),
                    keyboardType: TextInputType.number,
                    validator: (v) {
                      final n = int.tryParse(v ?? '');
                      if (n == null || n < 1) return '≥ 1';
                      return null;
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _timeLimitCtrl,
                    decoration:
                        const InputDecoration(labelText: 'Лимит (мин)'),
                    keyboardType: TextInputType.number,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            FilledButton(onPressed: _submit, child: const Text('Создать')),
          ],
        ),
      ),
    );
  }
}
