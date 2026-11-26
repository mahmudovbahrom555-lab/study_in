import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/localization/l10n.dart';
import '../providers/auth_provider.dart';

/// Тот же формат, что проверяет сервер: латиница, цифры и _, 4–32 символа,
/// первая — буква.
final _usernameRe = RegExp(r'^[a-z][a-z0-9_]{3,31}$');

/// @username репетитора: по нему ученики находят его открытые группы.
/// Пустое поле убирает username.
class UsernameDialog extends ConsumerStatefulWidget {
  const UsernameDialog({super.key, required this.initial});

  final String initial;

  @override
  ConsumerState<UsernameDialog> createState() => _UsernameDialogState();
}

class _UsernameDialogState extends ConsumerState<UsernameDialog> {
  late final _ctrl = TextEditingController(text: widget.initial);
  String? _error;
  bool _busy = false;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (_busy) return;
    final value = _ctrl.text.trim().replaceFirst('@', '').toLowerCase();
    if (value.isNotEmpty && !_usernameRe.hasMatch(value)) {
      setState(() => _error = context.l10n.usernameRules);
      return;
    }
    setState(() {
      _busy = true;
      _error = null;
    });
    final err = await ref.read(authProvider.notifier).setUsername(value);
    if (!mounted) return;
    if (err != null) {
      setState(() {
        _busy = false;
        _error = err;
      });
      return;
    }
    Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return AlertDialog(
      title: Text(l10n.usernameLabel),
      content: TextField(
        controller: _ctrl,
        autofocus: true,
        maxLength: 32,
        autocorrect: false,
        decoration: InputDecoration(
          prefixText: '@',
          helperText: l10n.usernameHint,
          helperMaxLines: 3,
          errorText: _error,
          errorMaxLines: 3,
        ),
        onSubmitted: (_) => _save(),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(l10n.cancel),
        ),
        FilledButton(onPressed: _busy ? null : _save, child: Text(l10n.save)),
      ],
    );
  }
}
