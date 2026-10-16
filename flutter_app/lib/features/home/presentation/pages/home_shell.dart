import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/router/routes.dart';
import '../../../../core/widgets/brand_mark.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../groups/presentation/pages/groups_page.dart';
import '../../../notifications/presentation/providers/notifications_provider.dart';
import '../../../parents/presentation/pages/parents_page.dart';
import '../../../../core/localization/l10n.dart';

/// Вкладка главного экрана. Новая вкладка (чат, карта для родителя…) —
/// одна запись в [_tabs] с ролями, которым она видна.
class _HomeTab {
  const _HomeTab({
    required this.icon,
    required this.label,
    required this.page,
    this.roles,
  });

  final IconData icon;
  final String Function(AppLocalizations) label;
  final Widget page;

  /// null — вкладка видна всем ролям.
  final Set<String>? roles;

  bool visible(String role) => roles == null || roles!.contains(role);
}

final _tabs = [
  _HomeTab(
    icon: LucideIcons.users,
    label: (l10n) => l10n.navGroups,
    page: const GroupsPage(),
  ),
  _HomeTab(
    icon: LucideIcons.heartHandshake,
    label: (l10n) => l10n.navChildren,
    page: const ParentsPage(),
    roles: {'parent'},
  ),
];

class HomeShell extends ConsumerStatefulWidget {
  const HomeShell({super.key});

  @override
  ConsumerState<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends ConsumerState<HomeShell> {
  int _tab = 0;

  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(notificationsProvider.notifier).load());
  }

  @override
  Widget build(BuildContext context) {
    final role = ref.watch(authProvider).user?.role ?? 'student';
    final tabs = [for (final t in _tabs) if (t.visible(role)) t];
    final index = _tab.clamp(0, tabs.length - 1);
    final unread = ref.watch(unreadCountProvider);

    return Scaffold(
      appBar: AppBar(
        title: const BrandTitle(),
        actions: [
          IconButton(
            tooltip: unread > 0 ? context.l10n.notificationsUnread(unread) : context.l10n.notifications,
            icon: Badge(
              isLabelVisible: unread > 0,
              label: Text('$unread'),
              child: const Icon(LucideIcons.bell),
            ),
            onPressed: () => context.push(Routes.notifications),
          ),
          IconButton(
            tooltip: context.l10n.account,
            icon: const Icon(LucideIcons.circleUser),
            onPressed: () => context.push(Routes.account),
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: IndexedStack(
        index: index,
        children: [for (final t in tabs) t.page],
      ),
      // Одна вкладка — нижняя панель не нужна.
      bottomNavigationBar: tabs.length > 1
          ? NavigationBar(
              selectedIndex: index,
              onDestinationSelected: (i) => setState(() => _tab = i),
              destinations: [
                for (final t in tabs)
                  NavigationDestination(
                    icon: Icon(t.icon),
                    label: t.label(context.l10n),
                  ),
              ],
            )
          : null,
    );
  }
}
