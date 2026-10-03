import 'package:flutter/material.dart';

import '../../domain/entities/group.dart';
import '../../../../core/localization/l10n.dart';

class GroupCard extends StatelessWidget {
  const GroupCard({super.key, required this.group, required this.onTap});

  final Group group;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: ListTile(
        onTap: onTap,
        title: Text(
          group.name,
          style: theme.textTheme.titleMedium,
        ),
        subtitle: group.subject != null ? Text(group.subject!) : null,
        trailing: group.isArchived
            ? Chip(
                label: Text(context.l10n.archivedChip),
                visualDensity: VisualDensity.compact,
                backgroundColor: theme.colorScheme.surfaceContainerHighest,
              )
            : null,
        leading: CircleAvatar(
          backgroundColor: theme.colorScheme.primaryContainer,
          child: Text(
            group.name.substring(0, 1).toUpperCase(),
            style: TextStyle(color: theme.colorScheme.onPrimaryContainer),
          ),
        ),
      ),
    );
  }
}
