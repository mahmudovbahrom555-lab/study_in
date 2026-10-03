import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../providers/auth_provider.dart';
import '../../../../core/localization/l10n.dart';

class RoleSelectPage extends ConsumerWidget {
  const RoleSelectPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isLoading = ref.watch(authProvider).isLoading;

    Future<void> selectRole(String role) async {
      await ref.read(authProvider.notifier).setRole(role);
      if (!context.mounted) return;
      final error = ref.read(authProvider).error;
      if (error != null) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(error)));
        return;
      }
      context.go(Routes.profileSetup);
    }

    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 48),
              Text(
                context.l10n.whoAreYou,
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              const SizedBox(height: 8),
              Text(
                context.l10n.roleWarning,
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: Colors.grey),
              ),
              const SizedBox(height: 48),
              _RoleCard(
                icon: Icons.school,
                title: context.l10n.roleTeacher,
                subtitle: context.l10n.roleTeacherDesc,
                onTap: isLoading ? null : () => selectRole('teacher'),
              ),
              const SizedBox(height: 16),
              _RoleCard(
                icon: Icons.person,
                title: context.l10n.roleStudent,
                subtitle: context.l10n.roleStudentDesc,
                onTap: isLoading ? null : () => selectRole('student'),
              ),
              const SizedBox(height: 16),
              _RoleCard(
                icon: Icons.family_restroom,
                title: context.l10n.roleParent,
                subtitle: context.l10n.roleParentDesc,
                onTap: isLoading ? null : () => selectRole('parent'),
              ),
              if (isLoading) ...[
                const SizedBox(height: 24),
                const Center(child: CircularProgressIndicator()),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _RoleCard extends StatelessWidget {
  const _RoleCard({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: Icon(icon, size: 32),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600)),
        subtitle: Text(subtitle),
        trailing: const Icon(Icons.chevron_right),
        onTap: onTap,
      ),
    );
  }
}
