import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';

import '../../../../core/config/app_config.dart';
import '../../../../core/localization/l10n.dart';
import '../../domain/entities/group.dart';

/// QR-код ссылки-приглашения: репетитор показывает его на занятии,
/// ученик наводит камеру и попадает на карточку группы с заявкой.
class GroupQrDialog extends StatelessWidget {
  const GroupQrDialog({super.key, required this.group});

  final Group group;

  @override
  Widget build(BuildContext context) {
    final link = AppConfig.joinLink(group.inviteCode);
    return AlertDialog(
      title: Text(group.name, textAlign: TextAlign.center),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Белый фон и в тёмной теме: камеры плохо читают светлый код на тёмном.
          Container(
            color: Colors.white,
            padding: const EdgeInsets.all(12),
            child: QrImageView(data: link, size: 220),
          ),
          const SizedBox(height: 12),
          Text(context.l10n.groupQrHint, textAlign: TextAlign.center),
          const SizedBox(height: 8),
          SelectableText(
            context.l10n.inviteCode(group.inviteCode),
            style: const TextStyle(fontWeight: FontWeight.w600),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(context.l10n.close),
        ),
      ],
    );
  }
}
