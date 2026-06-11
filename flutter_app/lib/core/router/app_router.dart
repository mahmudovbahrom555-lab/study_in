import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../features/auth/presentation/pages/splash_page.dart';
import '../../features/auth/presentation/pages/phone_page.dart';
import '../../features/auth/presentation/pages/verify_page.dart';
import '../../features/auth/presentation/pages/role_select_page.dart';
import '../../features/auth/presentation/pages/profile_setup_page.dart';
import '../../features/auth/presentation/providers/auth_provider.dart';
import '../../features/feed/presentation/pages/feed_page.dart';
import '../../features/groups/presentation/pages/groups_page.dart';
import '../../features/groups/presentation/pages/group_detail_page.dart';
import '../../features/quizzes/presentation/pages/quizzes_page.dart';
import '../../features/quizzes/presentation/pages/quiz_detail_page.dart';
import '../../features/quizzes/presentation/pages/quiz_attempt_page.dart';
import 'routes.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: Routes.splash,
    redirect: (context, state) {
      final isAuth = authState.isAuthenticated;
      final isLoading = authState.isLoading;
      final loc = state.matchedLocation;

      if (isLoading) return null;

      final publicRoutes = {Routes.splash, Routes.phone, Routes.verify};
      if (!isAuth && !publicRoutes.contains(loc)) return Routes.phone;

      return null;
    },
    routes: [
      GoRoute(
        path: Routes.splash,
        builder: (_, __) => const SplashPage(),
      ),
      GoRoute(
        path: Routes.phone,
        builder: (_, __) => const PhonePage(),
      ),
      GoRoute(
        path: Routes.verify,
        builder: (_, state) {
          final phone = state.extra as String? ?? '';
          return VerifyPage(phone: phone);
        },
      ),
      GoRoute(
        path: Routes.roleSelect,
        builder: (_, __) => const RoleSelectPage(),
      ),
      GoRoute(
        path: Routes.profileSetup,
        builder: (_, __) => const ProfileSetupPage(),
      ),
      GoRoute(
        path: Routes.home,
        builder: (_, __) => const GroupsPage(),
      ),
      GoRoute(
        path: Routes.groups,
        builder: (_, __) => const GroupsPage(),
      ),
      GoRoute(
        path: '/groups/:id',
        builder: (_, state) =>
            GroupDetailPage(groupId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/groups/:id/feed',
        builder: (_, state) =>
            FeedPage(groupId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/groups/:id/quizzes',
        builder: (_, state) =>
            QuizzesPage(groupId: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/groups/:groupId/quizzes/:quizId',
        builder: (_, state) => QuizDetailPage(
          groupId: state.pathParameters['groupId']!,
          quizId: state.pathParameters['quizId']!,
        ),
      ),
      GoRoute(
        path: '/groups/:groupId/quizzes/:quizId/attempt',
        builder: (_, state) => QuizAttemptPage(
          groupId: state.pathParameters['groupId']!,
          quizId: state.pathParameters['quizId']!,
        ),
      ),
    ],
  );
});
