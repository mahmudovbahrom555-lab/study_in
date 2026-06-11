import 'package:flutter/material.dart';

import '../../domain/entities/group.dart';

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
                label: const Text('Архив'),
                visualDensity: VisualDensity.compact,
                backgroundColor:
                    theme.colorScheme.surfaceVariant,
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
