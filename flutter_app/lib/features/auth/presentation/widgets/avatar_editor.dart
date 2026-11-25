import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/widgets/user_avatar.dart';
import '../providers/auth_provider.dart';

/// Выбор аватара. Ученики — только животные (фото детей не храним),
/// репетитор и родитель — ещё и своё фото.
Future<void> showAvatarEditor(BuildContext context) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    builder: (_) => const _AvatarEditor(),
  );
}

class _AvatarEditor extends ConsumerStatefulWidget {
  const _AvatarEditor();

  @override
  ConsumerState<_AvatarEditor> createState() => _AvatarEditorState();
}

class _AvatarEditorState extends ConsumerState<_AvatarEditor> {
  bool _busy = false;

  AuthNotifier get _auth => ref.read(authProvider.notifier);

  Future<void> _run(Future<String?> pending) async {
    final messenger = ScaffoldMessenger.of(context);
    final navigator = Navigator.of(context);
    setState(() => _busy = true);
    final error = await pending;
    if (!mounted) return;
    setState(() => _busy = false);
    if (error != null) {
      messenger.showSnackBar(SnackBar(content: Text(error)));
      return;
    }
    navigator.pop();
  }

  Future<void> _uploadPhoto() async {
    // Сжатие на устройстве: аватару хватает 512 px, на сервере лимит 2 МБ.
    final x = await ImagePicker().pickImage(
      source: ImageSource.gallery,
      maxWidth: 512,
      maxHeight: 512,
      imageQuality: 85,
    );
    if (x == null) return;
    final path = kIsWeb ? null : x.path;
    final bytes = kIsWeb ? await x.readAsBytes() : null;
    await _run(_auth.uploadAvatar(path: path, bytes: bytes, name: x.name));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final user = ref.watch(authProvider).user;
    final hasAvatar = user?.avatarUrl != null || user?.avatarPreset != null;

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.avatarTitle,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            if (_busy) const LinearProgressIndicator(),
            if (user != null && !user.isStudent)
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const Icon(LucideIcons.image),
                title: Text(l10n.avatarUploadPhoto),
                onTap: _busy ? null : _uploadPhoto,
              ),
            Text(l10n.avatarAnimals),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final preset in avatarPresets.keys)
                  InkWell(
                    customBorder: const CircleBorder(),
                    onTap: _busy
                        ? null
                        : () => _run(_auth.setAvatarPreset(preset)),
                    child: UserAvatar(
                      name: '',
                      avatarPreset: preset,
                      radius: 26,
                    ),
                  ),
              ],
            ),
            if (hasAvatar) ...[
              const SizedBox(height: 8),
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const Icon(LucideIcons.trash2),
                title: Text(l10n.avatarRemove),
                onTap: _busy ? null : () => _run(_auth.deleteAvatar()),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
