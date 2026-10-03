import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/feed_provider.dart';
import '../../../../core/localization/l10n.dart';

class CreatePostSheet extends ConsumerStatefulWidget {
  const CreatePostSheet({super.key, required this.groupId});

  final String groupId;

  @override
  ConsumerState<CreatePostSheet> createState() => _CreatePostSheetState();
}

class _CreatePostSheetState extends ConsumerState<CreatePostSheet> {
  final _bodyCtrl = TextEditingController();
  bool _pinned = false;

  @override
  void dispose() {
    _bodyCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isLoading = ref.watch(feedProvider(widget.groupId)).isLoading;

    return Padding(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 16,
        bottom: MediaQuery.of(context).viewInsets.bottom + 16,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            context.l10n.newPublication,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _bodyCtrl,
            decoration: InputDecoration(
              labelText: context.l10n.messageText,
              alignLabelWithHint: true,
            ),
            maxLines: 5,
            textCapitalization: TextCapitalization.sentences,
          ),
          const SizedBox(height: 8),
          SwitchListTile(
            title: Text(context.l10n.pin),
            value: _pinned,
            onChanged: (v) => setState(() => _pinned = v),
            contentPadding: EdgeInsets.zero,
          ),
          const SizedBox(height: 12),
          ElevatedButton(
            onPressed: isLoading ? null : _submit,
            child: isLoading
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Text(context.l10n.publish),
          ),
        ],
      ),
    );
  }

  Future<void> _submit() async {
    final body = _bodyCtrl.text.trim();
    if (body.isEmpty) return;
    await ref
        .read(feedProvider(widget.groupId).notifier)
        .createPost(body, pinned: _pinned);
    if (mounted) Navigator.of(context).pop();
  }
}
