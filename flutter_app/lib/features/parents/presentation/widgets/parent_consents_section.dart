import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/network/error_message.dart';
import '../../../groups/domain/entities/join_request.dart';
import '../../../groups/presentation/providers/groups_provider.dart';

/// Заявки детей младше 13 в группы: репетитор увидит заявку, только когда
/// родитель её разрешит. Пока заявок нет, раздел не занимает места.
class ParentConsentsSection extends ConsumerWidget {
  const ParentConsentsSection({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final consents = ref.watch(parentConsentsProvider).valueOrNull;
    if (consents == null || consents.isEmpty) return const SizedBox.shrink();

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Text(
                context.l10n.consentsTitle,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
            for (final c in consents) _ConsentTile(consent: c),
          ],
        ),
      ),
    );
  }
}

class _ConsentTile extends ConsumerWidget {
  const _ConsentTile({required this.consent});

  final ParentConsent consent;

  Future<void> _decide(
    BuildContext context,
    WidgetRef ref, {
    required bool approve,
  }) async {
    final messenger = ScaffoldMessenger.of(context);
    try {
      final repo = ref.read(groupsRepositoryProvider);
      await repo.decideParentConsent(consent.id, approve: approve);
    } catch (e) {
      messenger.showSnackBar(SnackBar(content: Text(userErrorMessage(e))));
    }
    ref.invalidate(parentConsentsProvider);
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = context.l10n;
    final c = consent;
    final subject = c.subject;
    final group = subject == null ? c.groupName : '${c.groupName} · $subject';
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(l10n.consentAsk(c.studentName, group, c.teacherName)),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              TextButton(
                onPressed: () => _decide(context, ref, approve: false),
                child: Text(l10n.consentDecline),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: () => _decide(context, ref, approve: true),
                child: Text(l10n.consentAllow),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
