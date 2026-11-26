import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/widgets/user_avatar.dart';
import '../../domain/entities/join_request.dart';
import '../providers/groups_provider.dart';

/// Заявки учеников в группу — видит только её репетитор.
/// Пока заявок нет, раздел не занимает места.
class JoinRequestsSection extends ConsumerWidget {
  const JoinRequestsSection({super.key, required this.groupId});

  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final requests = ref.watch(joinRequestsProvider(groupId)).valueOrNull;
    if (requests == null || requests.isEmpty) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: Text(
            context.l10n.joinRequests(requests.length),
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        for (final r in requests)
          _RequestTile(
            request: r,
            onDecide: (accept) => _decide(context, ref, r, accept),
          ),
        const Divider(height: 1),
      ],
    );
  }

  Future<void> _decide(
    BuildContext context,
    WidgetRef ref,
    JoinRequestItem request,
    bool accept,
  ) async {
    final messenger = ScaffoldMessenger.of(context);
    try {
      final repo = ref.read(groupsRepositoryProvider);
      await repo.decideJoinRequest(groupId, request.id, accept: accept);
      if (accept) ref.read(membersProvider(groupId).notifier).load();
    } catch (e) {
      messenger.showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
    }
    ref.invalidate(joinRequestsProvider(groupId));
  }
}

class _RequestTile extends StatelessWidget {
  const _RequestTile({required this.request, required this.onDecide});

  final JoinRequestItem request;
  final void Function(bool accept) onDecide;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: UserAvatar(
        name: request.studentName,
        avatarUrl: request.avatarUrl,
        avatarPreset: request.avatarPreset,
      ),
      title: Text(request.studentName),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            tooltip: context.l10n.declineRequest,
            icon: const Icon(LucideIcons.x),
            onPressed: () => onDecide(false),
          ),
          IconButton(
            tooltip: context.l10n.accept,
            icon: const Icon(LucideIcons.check),
            color: Theme.of(context).colorScheme.primary,
            onPressed: () => onDecide(true),
          ),
        ],
      ),
    );
  }
}
