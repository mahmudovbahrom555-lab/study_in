import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../groups/presentation/pages/groups_page.dart';
import '../../../notifications/presentation/providers/notifications_provider.dart';
import '../../../parents/presentation/pages/parents_page.dart';
import '../../../../core/localization/l10n.dart';

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
    final isParent = role == 'parent';
    final unread = ref.watch(unreadCountProvider);

    final pages = isParent
        ? const [GroupsPage(), ParentsPage()]
        : const [GroupsPage()];

    return Scaffold(
      appBar: AppBar(
        title: Text(context.l10n.appName),
        actions: [
          IconButton(
            tooltip: context.l10n.account,
            icon: const Icon(Icons.account_circle_outlined),
            onPressed: () => context.push(Routes.account),
          ),
          IconButton(
            tooltip: unread > 0 ? context.l10n.notificationsUnread(unread) : context.l10n.notifications,
            icon: Badge(
              isLabelVisible: unread > 0,
              label: Text('$unread'),
              child: const Icon(Icons.notifications_outlined),
            ),
            onPressed: () => context.push(Routes.notifications),
          ),
        ],
      ),
      body: IndexedStack(
        index: _tab.clamp(0, pages.length - 1),
        children: pages,
      ),
      bottomNavigationBar: isParent
          ? NavigationBar(
              selectedIndex: _tab,
              onDestinationSelected: (i) => setState(() => _tab = i),
              destinations: [
                NavigationDestination(
                  icon: const Icon(Icons.groups_outlined),
                  selectedIcon: const Icon(Icons.groups),
                  label: context.l10n.navGroups,
                ),
                NavigationDestination(
                  icon: const Icon(Icons.child_care_outlined),
                  selectedIcon: const Icon(Icons.child_care),
                  label: context.l10n.navChildren,
                ),
              ],
            )
          : null,
    );
  }
}
