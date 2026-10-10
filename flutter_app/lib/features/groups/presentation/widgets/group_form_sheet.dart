import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/group.dart';
import '../providers/groups_provider.dart';
import '../../../../core/localization/l10n.dart';

/// Создание группы, а с [group] — её редактирование.
class GroupFormSheet extends ConsumerStatefulWidget {
  const GroupFormSheet({super.key, this.group});

  final Group? group;

  @override
  ConsumerState<GroupFormSheet> createState() => _GroupFormSheetState();
}

class _GroupFormSheetState extends ConsumerState<GroupFormSheet> {
  final _formKey = GlobalKey<FormState>();
  late final _nameCtrl = TextEditingController(text: widget.group?.name);
  late final _subjectCtrl = TextEditingController(text: widget.group?.subject);
  late final _descCtrl = TextEditingController(text: widget.group?.description);

  bool get _isEdit => widget.group != null;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _subjectCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isLoading = ref.watch(groupsProvider).isLoading;

    return Padding(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 16,
        bottom: MediaQuery.of(context).viewInsets.bottom + 16,
      ),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              _isEdit ? context.l10n.editGroup : context.l10n.newGroup,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _nameCtrl,
              decoration: InputDecoration(labelText: context.l10n.titleRequired),
              validator: (v) => (v == null || v.trim().length < 2)
                  ? context.l10n.minTwoChars
                  : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _subjectCtrl,
              decoration: InputDecoration(labelText: context.l10n.subject),
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _descCtrl,
              decoration: InputDecoration(labelText: context.l10n.description),
              maxLines: 2,
            ),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: isLoading ? null : _submit,
              child: isLoading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(_isEdit ? context.l10n.save : context.l10n.create),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final notifier = ref.read(groupsProvider.notifier);
    final name = _nameCtrl.text.trim();
    final subject = _subjectCtrl.text.trim();
    final description = _descCtrl.text.trim();
    if (_isEdit) {
      // Пустая строка очищает поле на сервере.
      await notifier.updateGroup(
        widget.group!.id,
        name: name,
        subject: subject,
        description: description,
      );
    } else {
      await notifier.createGroup(
        name: name,
        subject: subject.isEmpty ? null : subject,
        description: description.isEmpty ? null : description,
      );
    }
    if (mounted) Navigator.of(context).pop();
  }
}
