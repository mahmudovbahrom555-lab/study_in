import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/router/routes.dart';
import '../../../../core/widgets/user_avatar.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/join_request.dart';
import '../providers/groups_provider.dart';

/// Карточка группы по ссылке-приглашению: ученик отправляет заявку,
/// в группу он попадёт, когда репетитор её примет.
class JoinGroupPage extends ConsumerStatefulWidget {
  const JoinGroupPage({super.key, required this.code});

  final String code;

  @override
  ConsumerState<JoinGroupPage> createState() => _JoinGroupPageState();
}

class _JoinGroupPageState extends ConsumerState<JoinGroupPage> {
  bool _sending = false;
  JoinRequest? _sent;

  @override
  Widget build(BuildContext context) {
    final code = widget.code.trim().toUpperCase();
    final preview = ref.watch(groupPreviewProvider(code));
    final isStudent = ref.watch(authProvider).user?.role == 'student';

    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.joinGroupTitle)),
      body: preview.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => _Message(text: context.l10n.joinLinkInvalid),
        data: (g) => ListView(
          padding: const EdgeInsets.all(24),
          children: [
            _PreviewCard(group: g),
            const SizedBox(height: 24),
            if (!isStudent)
              Text(context.l10n.joinOnlyStudents, textAlign: TextAlign.center)
            else if (_sent case final sent?)
              Text(
                joinRequestSentText(context, sent),
                textAlign: TextAlign.center,
              )
            else
              FilledButton(
                onPressed: _sending ? null : () => _send(code),
                child: Text(context.l10n.sendJoinRequest),
              ),
            const SizedBox(height: 12),
            TextButton(
              onPressed: () => context.go(Routes.home),
              child: Text(context.l10n.toMyGroups),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _send(String code) async {
    final messenger = ScaffoldMessenger.of(context);
    setState(() => _sending = true);
    try {
      final sent = await ref.read(groupsRepositoryProvider).joinGroup(code);
      ref.invalidate(myJoinRequestsProvider);
      if (mounted) setState(() => _sent = sent);
    } catch (e) {
      messenger.showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }
}

class _PreviewCard extends StatelessWidget {
  const _PreviewCard({required this.group});

  final GroupPreview group;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(group.name, style: theme.textTheme.titleLarge),
            if (group.subject != null) ...[
              const SizedBox(height: 4),
              Text(
                group.subject!,
                style: TextStyle(color: theme.colorScheme.primary),
              ),
            ],
            if (group.description != null) ...[
              const SizedBox(height: 8),
              Text(group.description!),
            ],
            const SizedBox(height: 16),
            Row(
              children: [
                UserAvatar(
                  name: group.teacherName,
                  avatarUrl: group.teacherAvatarUrl,
                  avatarPreset: group.teacherAvatarPreset,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(context.l10n.groupTutor(group.teacherName)),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(context.l10n.groupStudentsCount(group.memberCount)),
          ],
        ),
      ),
    );
  }
}

class _Message extends StatelessWidget {
  const _Message({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Text(text, textAlign: TextAlign.center),
      ),
    );
  }
}

/// Что сказать ученику после заявки: ребёнку младше 13 — что ждём родителя,
/// а если родитель не привязан — как его позвать.
String joinRequestSentText(BuildContext context, JoinRequest r) {
  final l10n = context.l10n;
  if (!r.waitsForParent) return l10n.joinRequestSent;
  return r.hasParent ? l10n.joinWaitsParent : l10n.joinNeedsParent;
}
