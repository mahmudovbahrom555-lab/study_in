import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../../core/app_restart.dart';
import '../../../../core/localization/l10n.dart';
import '../../../../core/localization/language_switcher.dart';
import '../../../../core/network/error_message.dart';
import '../../../parents/presentation/providers/parents_provider.dart';
import '../providers/auth_provider.dart';

/// Аккаунт: кто вошёл, язык интерфейса и выход.
class AccountPage extends ConsumerWidget {
  const AccountPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = context.l10n;
    final user = ref.watch(authProvider).user;

    return Scaffold(
      appBar: AppBar(title: Text(l10n.account)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ListTile(
            leading: const Icon(Icons.person_outline),
            title: Text(user?.name.isNotEmpty == true ? user!.name : '—'),
            subtitle: Text(l10n.nameLabel),
          ),
          ListTile(
            leading: const Icon(Icons.phone_outlined),
            title: Text(user?.phone ?? '—'),
            subtitle: Text(l10n.phoneNumber),
          ),
          ListTile(
            leading: const Icon(Icons.badge_outlined),
            title: Text(_roleName(l10n, user?.role)),
            subtitle: Text(l10n.roleLabel),
          ),
          if (user?.role == 'student') ...[
            const Divider(height: 32),
            const _ParentCodeSection(),
          ],
          const Divider(height: 32),
          ListTile(
            leading: const Icon(Icons.language),
            title: Text(l10n.language),
            trailing: const LanguageSwitcher(),
          ),
          const Divider(height: 32),
          ListTile(
            leading: Icon(Icons.logout, color: Theme.of(context).colorScheme.error),
            title: Text(
              l10n.signOut,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
            onTap: () => _confirmSignOut(context, ref),
          ),
        ],
      ),
    );
  }

  String _roleName(AppLocalizations l10n, String? role) => switch (role) {
        'teacher' => l10n.roleTeacher,
        'student' => l10n.roleStudent,
        'parent' => l10n.roleParent,
        _ => '—',
      };

  Future<void> _confirmSignOut(BuildContext context, WidgetRef ref) async {
    final l10n = context.l10n;
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.signOutQ),
        content: Text(l10n.signOutBody),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text(l10n.signOut),
          ),
        ],
      ),
    );
    if (ok != true || !context.mounted) return;

    await ref.read(authProvider.notifier).logout();
    if (!context.mounted) return;
    // Полный перезапуск: следующий аккаунт не увидит закэшированные
    // группы, оценки и уведомления предыдущего.
    AppRestartScope.restart(context);
  }
}

/// Ученик получает короткий код и передаёт его родителю — так родитель
/// привязывается с согласия ученика, без ввода длинного ID.
class _ParentCodeSection extends ConsumerStatefulWidget {
  const _ParentCodeSection();

  @override
  ConsumerState<_ParentCodeSection> createState() => _ParentCodeSectionState();
}

class _ParentCodeSectionState extends ConsumerState<_ParentCodeSection> {
  ({String code, DateTime expiresAt})? _code;
  String? _error;
  bool _loading = false;

  Future<void> _generate() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final code = await ref.read(parentsRepositoryProvider).createLinkCode();
      if (mounted) setState(() => _code = code);
    } catch (e) {
      if (mounted) setState(() => _error = userErrorMessage(e));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _copy() async {
    await Clipboard.setData(ClipboardData(text: _code!.code));
    if (!mounted) return;
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(context.l10n.copied)));
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final theme = Theme.of(context);
    final code = _code;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.family_restroom),
              const SizedBox(width: 16),
              Text(l10n.parentCodeTitle, style: theme.textTheme.titleMedium),
            ],
          ),
          const SizedBox(height: 8),
          Text(l10n.parentCodeHelp, style: theme.textTheme.bodySmall),
          const SizedBox(height: 12),
          if (code != null) ...[
            Row(
              children: [
                SelectableText(
                  code.code,
                  style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                    letterSpacing: 4,
                  ),
                ),
                const SizedBox(width: 8),
                IconButton(
                  tooltip: l10n.copy,
                  icon: const Icon(Icons.copy),
                  onPressed: _copy,
                ),
              ],
            ),
            Text(
              l10n.codeValidUntil(
                DateFormat.MMMd(l10n.localeName).add_Hm().format(code.expiresAt),
              ),
              style: theme.textTheme.bodySmall,
            ),
            const SizedBox(height: 8),
          ],
          if (_error != null) ...[
            Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
            const SizedBox(height: 8),
          ],
          OutlinedButton(
            onPressed: _loading ? null : _generate,
            child: Text(code == null ? l10n.getCode : l10n.newCode),
          ),
        ],
      ),
    );
  }
}
