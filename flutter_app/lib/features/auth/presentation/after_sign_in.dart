import 'package:flutter/widgets.dart';
import 'package:go_router/go_router.dart';

import '../../../core/router/routes.dart';
import '../domain/entities/user.dart';

/// Следующий экран после входа — одинаково для SMS и Google.
void goAfterSignIn(
  BuildContext context,
  User? user, {
  required bool isNewUser,
}) {
  // Роль могла не сохраниться при прошлом входе — без неё API отвечает 403.
  if (isNewUser || (user != null && !user.hasRole)) {
    context.go(Routes.roleSelect);
  } else if (user != null && user.name.isEmpty) {
    context.go(Routes.profileSetup);
  } else {
    context.go(Routes.home);
  }
}
