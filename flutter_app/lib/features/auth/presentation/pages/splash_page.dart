import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../../../core/widgets/error_view.dart';
import '../providers/auth_provider.dart';

class SplashPage extends ConsumerStatefulWidget {
  const SplashPage({super.key});

  @override
  ConsumerState<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends ConsumerState<SplashPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _init());
  }

  Future<void> _init() async {
    await ref.read(authProvider.notifier).checkAuth();
    if (!mounted) return;

    final state = ref.read(authProvider);
    if (state.isAuthenticated) {
      final user = state.user;
      if (user != null && !user.hasRole) {
        context.go(Routes.roleSelect);
      } else if (user != null && user.name.isEmpty) {
        context.go(Routes.profileSetup);
      } else {
        context.go(Routes.home);
      }
    } else if (state.error == null) {
      context.go(Routes.phone);
    }
    // Иначе сервер недоступен, а вход сохранён — остаёмся здесь с «Повторить».
  }

  @override
  Widget build(BuildContext context) {
    final auth = ref.watch(authProvider);
    return Scaffold(
      body: !auth.isLoading && auth.error != null
          ? ErrorView(message: auth.error, onRetry: _init)
          : const Center(child: CircularProgressIndicator()),
    );
  }
}
