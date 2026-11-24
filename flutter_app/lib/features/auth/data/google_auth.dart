import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';

import '../../../core/config/app_config.dart';

/// Вход через Google (google_sign_in 7). Результат — ID-токен, который
/// проверяет сервер (POST /auth/google).
///
/// На Android вход запускает своя кнопка через [signIn]. На вебе кнопку рисует
/// Google (GIS), а токен приходит через [webIdTokens].
class GoogleAuth {
  GoogleAuth._();

  static final instance = GoogleAuth._();

  Future<void>? _init;

  bool get isConfigured => AppConfig.googleClientId.isNotEmpty;

  /// initialize разрешён один раз за запуск приложения.
  Future<void> ensureInitialized() {
    return _init ??= GoogleSignIn.instance.initialize(
      // На вебе нужен clientId; на Android — serverClientId, чтобы audience
      // ID-токена совпал с GOOGLE_CLIENT_ID на сервере.
      clientId: kIsWeb ? AppConfig.googleClientId : null,
      serverClientId: kIsWeb ? null : AppConfig.googleClientId,
    );
  }

  /// Android: системное окно выбора аккаунта. null — человек закрыл окно.
  Future<String?> signIn() async {
    await ensureInitialized();
    try {
      final account = await GoogleSignIn.instance.authenticate();
      return account.authentication.idToken;
    } on GoogleSignInException catch (e) {
      if (e.code == GoogleSignInExceptionCode.canceled) return null;
      rethrow;
    }
  }

  /// Веб: ID-токены после нажатия на кнопку Google.
  Stream<String> get webIdTokens async* {
    await for (final e in GoogleSignIn.instance.authenticationEvents) {
      if (e is GoogleSignInAuthenticationEventSignIn) {
        final token = e.user.authentication.idToken;
        if (token != null) yield token;
      }
    }
  }
}
