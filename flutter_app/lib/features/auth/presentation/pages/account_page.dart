import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/app_restart.dart';
import '../../../../core/config/app_config.dart';
import '../../../../core/localization/l10n.dart';
import '../../../../core/localization/language_switcher.dart';
import '../../../../core/network/error_message.dart';
import '../../../groups/presentation/providers/groups_provider.dart';
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
            leading: const Icon(LucideIcons.user),
            title: Text(user?.name.isNotEmpty == true ? user!.name : '—'),
            subtitle: Text(l10n.nameLabel),
          ),
          ListTile(
            leading: const Icon(LucideIcons.phone),
            title: Text(user?.phone ?? '—'),
            subtitle: Text(l10n.phoneNumber),
          ),
          ListTile(
            leading: const Icon(LucideIcons.idCard),
            title: Text(_roleName(l10n, user?.role)),
            subtitle: Text(l10n.roleLabel),
          ),
          if (user?.role == 'student') ...[
            const Divider(height: 32),
            const _ParentCodeSection(),
          ],
          const Divider(height: 32),
          ListTile(
            leading: const Icon(LucideIcons.languages),
            title: Text(l10n.language),
            trailing: const LanguageSwitcher(),
          ),
          ListTile(
            leading: const Icon(LucideIcons.shieldCheck),
            title: Text(l10n.privacyPolicy),
            trailing: const Icon(LucideIcons.externalLink, size: 18),
            onTap: () => launchUrl(
              Uri.parse(AppConfig.privacyUrl),
              mode: LaunchMode.externalApplication,
            ),
          ),
          const Divider(height: 32),
          ListTile(
            leading: Icon(LucideIcons.logOut, color: Theme.of(context).colorScheme.error),
            title: Text(
              l10n.signOut,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
            onTap: () => _confirmSignOut(context, ref),
          ),
          // Требование Google Play: удалить аккаунт можно прямо в приложении.
          ListTile(
            leading: Icon(LucideIcons.trash2, color: Theme.of(context).colorScheme.error),
            title: Text(
              l10n.deleteAccount,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
            onTap: () => _confirmDelete(context, ref),
          ),
        ],
      ),
    );
  }

  Future<void> _confirmDelete(BuildContext context, WidgetRef ref) async {
    final l10n = context.l10n;
    final user = ref.read(authProvider).user;
    // Репетитор должен видеть, что вместе с ним исчезнут его группы.
    final ownGroups = user?.role == 'teacher'
        ? ref.read(groupsProvider).groups.where((g) => g.teacherId == user!.id).length
        : 0;

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) {
        final error = Theme.of(ctx).colorScheme.error;
        return AlertDialog(
          title: Text(l10n.deleteAccountQ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(l10n.deleteAccountBody),
              if (ownGroups > 0) ...[
                const SizedBox(height: 12),
                Text(
                  l10n.deleteAccountGroups(ownGroups),
                  style: TextStyle(color: error, fontWeight: FontWeight.w600),
                ),
              ],
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text(l10n.cancel),
            ),
            FilledButton(
              style: FilledButton.styleFrom(backgroundColor: error),
              onPressed: () => Navigator.pop(ctx, true),
              child: Text(l10n.deleteAccountForever),
            ),
          ],
        );
      },
    );
    if (ok != true || !context.mounted) return;

    showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (_) => const Center(child: CircularProgressIndicator()),
    );
    final err = await ref.read(authProvider.notifier).deleteAccount();
    if (!context.mounted) return;
    Navigator.of(context).pop(); // индикатор
    if (err != null) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      return;
    }
    // Как при выходе: следующий аккаунт не должен увидеть чужие данные из кэша.
    AppRestartScope.restart(context);
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
              const Icon(LucideIcons.heartHandshake),
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
                  icon: const Icon(LucideIcons.copy),
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
