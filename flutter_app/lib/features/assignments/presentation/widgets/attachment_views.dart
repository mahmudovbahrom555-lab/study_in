import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/localization/l10n.dart';
import '../../domain/entities/submission.dart';

/// Фото открываем во весь экран внутри приложения, PDF и Word — во внешнем
/// просмотрщике телефона (ссылка подписана и живёт час).
Future<void> openAttachment(BuildContext context, AttachedFile file) async {
  if (file.isImage) {
    await Navigator.of(context).push(
      MaterialPageRoute<void>(builder: (_) => _FullScreenImage(file: file)),
    );
    return;
  }
  final opened = await launchUrl(
    Uri.parse(file.url),
    mode: LaunchMode.externalApplication,
  );
  if (!opened && context.mounted) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(context.l10n.hwCantOpenFile)));
  }
}

/// Квадратная миниатюра: фото — превью, остальное — значок документа.
class AttachmentThumb extends StatelessWidget {
  const AttachmentThumb({super.key, required this.child, this.size = 48});

  /// Image.file / Image.network или null для документа.
  final Widget? child;
  final double size;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: Container(
        width: size,
        height: size,
        color: theme.colorScheme.surfaceContainerHighest,
        alignment: Alignment.center,
        child: child ??
            Icon(LucideIcons.fileText, color: theme.colorScheme.primary),
      ),
    );
  }
}

Widget networkPreview(AttachedFile f) => Image.network(
      f.url,
      fit: BoxFit.cover,
      width: double.infinity,
      height: double.infinity,
      errorBuilder: (_, __, ___) => const Icon(LucideIcons.imageOff),
    );

/// Файл сдачи у репетитора: фото — миниатюрой, документ — чипом с именем.
class AttachmentChip extends StatelessWidget {
  const AttachmentChip({super.key, required this.file});

  final AttachedFile file;

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    if (file.isImage) {
      return Semantics(
        button: true,
        label: l10n.hwOpenFile(file.name),
        child: InkWell(
          borderRadius: BorderRadius.circular(8),
          onTap: () => openAttachment(context, file),
          child: AttachmentThumb(size: 72, child: networkPreview(file)),
        ),
      );
    }
    return ActionChip(
      avatar: const Icon(LucideIcons.fileText, size: 18),
      label: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 200),
        child: Text(file.name, overflow: TextOverflow.ellipsis),
      ),
      tooltip: l10n.hwOpenFile(file.name),
      onPressed: () => openAttachment(context, file),
    );
  }
}

class _FullScreenImage extends StatelessWidget {
  const _FullScreenImage({required this.file});

  final AttachedFile file;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        title: Text(file.name, overflow: TextOverflow.ellipsis),
      ),
      body: InteractiveViewer(
        maxScale: 5,
        child: Center(
          child: Image.network(
            file.url,
            loadingBuilder: (_, child, progress) => progress == null
                ? child
                : const CircularProgressIndicator(color: Colors.white),
            errorBuilder: (_, __, ___) =>
                const Icon(LucideIcons.imageOff, color: Colors.white),
          ),
        ),
      ),
    );
  }
}
