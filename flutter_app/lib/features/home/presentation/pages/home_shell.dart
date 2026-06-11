import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../groups/presentation/pages/groups_page.dart';
import '../../../parents/presentation/pages/parents_page.dart';

class HomeShell extends ConsumerStatefulWidget {
  const HomeShell({super.key});

  @override
  ConsumerState<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends ConsumerState<HomeShell> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    final role = ref.watch(authProvider).user?.role ?? 'student';
    final isParent = role == 'parent';

    final pages = isParent
        ? const [GroupsPage(), ParentsPage()]
        : const [GroupsPage()];

    return Scaffold(
      body: IndexedStack(
        index: _tab.clamp(0, pages.length - 1),
        children: pages,
      ),
      bottomNavigationBar: isParent
          ? NavigationBar(
              selectedIndex: _tab,
              onDestinationSelected: (i) => setState(() => _tab = i),
              destinations: const [
                NavigationDestination(
                  icon: Icon(Icons.groups_outlined),
                  selectedIcon: Icon(Icons.groups),
                  label: 'Группы',
                ),
                NavigationDestination(
                  icon: Icon(Icons.child_care_outlined),
                  selectedIcon: Icon(Icons.child_care),
                  label: 'Дети',
                ),
              ],
            )
          : null,
    );
  }
}
