import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/localization/l10n.dart';
import '../providers/groups_provider.dart';

/// Заявки ученика, которые репетитор ещё не принял.
class PendingRequestsBanner extends ConsumerWidget {
  const PendingRequestsBanner({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final requests = ref.watch(myJoinRequestsProvider).valueOrNull;
    if (requests == null || requests.isEmpty) return const SizedBox.shrink();

    return Card(
      margin: const EdgeInsets.fromLTRB(16, 16, 16, 0),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Text(
                context.l10n.pendingJoinRequests,
                style: Theme.of(context).textTheme.titleSmall,
              ),
            ),
            for (final r in requests)
              ListTile(
                dense: true,
                leading: const Icon(LucideIcons.clock),
                title: Text(r.groupName),
                subtitle: r.waitsForParent
                    ? Text(context.l10n.waitsParentShort)
                    : null,
              ),
          ],
        ),
      ),
    );
  }
}
