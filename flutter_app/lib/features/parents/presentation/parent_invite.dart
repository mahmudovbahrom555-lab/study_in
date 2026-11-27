import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/config/app_config.dart';
import '../../../core/localization/l10n.dart';
import '../../../core/network/error_message.dart';
import 'providers/parents_provider.dart';

/// Код из ссылки `mamu.education/#/parent/<код>`, открытой до входа:
/// главный экран родителя привяжет ребёнка, как только родитель войдёт.
String? pendingParentCode;

/// Ученик копирует ссылку-приглашение для родителя (код живёт 7 дней)
/// вместе с коротким текстом — остаётся вставить её в мессенджер.
Future<void> copyParentInvite(BuildContext context, WidgetRef ref) async {
  final l10n = context.l10n;
  final messenger = ScaffoldMessenger.of(context);
  try {
    final repo = ref.read(parentsRepositoryProvider);
    final code = await repo.createLinkCode(invite: true);
    final link = AppConfig.parentInviteLink(code.code);
    await Clipboard.setData(ClipboardData(text: l10n.parentInviteText(link)));
    messenger.showSnackBar(SnackBar(content: Text(l10n.parentInviteCopied)));
  } catch (e) {
    messenger.showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
  }
}
