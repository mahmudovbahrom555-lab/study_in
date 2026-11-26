import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/groups_provider.dart';
import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';

class JoinGroupSheet extends ConsumerStatefulWidget {
  const JoinGroupSheet({super.key});

  @override
  ConsumerState<JoinGroupSheet> createState() => _JoinGroupSheetState();
}

class _JoinGroupSheetState extends ConsumerState<JoinGroupSheet> {
  final _formKey = GlobalKey<FormState>();
  final _codeCtrl = TextEditingController();
  bool _sending = false;

  @override
  void dispose() {
    _codeCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
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
              context.l10n.joinGroupTitle,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _codeCtrl,
              decoration: InputDecoration(
                labelText: context.l10n.inviteCodeLabel,
                hintText: context.l10n.inviteCodeHint,
              ),
              textCapitalization: TextCapitalization.characters,
              validator: (v) => (v == null || v.trim().length < 4)
                  ? context.l10n.enterCode
                  : null,
            ),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: _sending ? null : _submit,
              child: _sending
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(context.l10n.join),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    final messenger = ScaffoldMessenger.of(context);
    final sent = context.l10n.joinRequestSent;
    final code = _codeCtrl.text.trim().toUpperCase();
    setState(() => _sending = true);
    try {
      await ref.read(groupsRepositoryProvider).joinGroup(code);
      ref.invalidate(myJoinRequestsProvider);
      messenger.showSnackBar(SnackBar(content: Text(sent)));
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      messenger.showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
      if (mounted) setState(() => _sending = false);
    }
  }
}
