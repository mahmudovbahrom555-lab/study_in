import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/localization/locale_provider.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/storage/secure_storage.dart';
import '../../data/auth_api.dart';
import '../../data/auth_repository_impl.dart';
import '../../data/session_tabs.dart';
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';

// ─── Providers ───────────────────────────────────────────────────────────────

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final dio = ref.watch(dioProvider);
  final storage = ref.watch(secureStorageProvider);
  return AuthRepositoryImpl(AuthApi(dio), storage);
});

// ─── State ───────────────────────────────────────────────────────────────────

class AuthState {
  const AuthState({
    this.user,
    this.isLoading = false,
    this.error,
    this.isAuthenticated = false,
  });

  final User? user;
  final bool isLoading;
  final String? error;
  final bool isAuthenticated;

  AuthState copyWith({
    User? user,
    bool? isLoading,
    String? error,
    bool? isAuthenticated,
    bool clearError = false,
  }) {
    return AuthState(
      user: user ?? this.user,
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : (error ?? this.error),
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
    );
  }
}

// ─── Notifier ─────────────────────────────────────────────────────────────────

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._repo, this._storage) : super(const AuthState());

  final AuthRepository _repo;
  final SecureStorage _storage;

  Future<void> checkAuth() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final hasToken = await _storage.hasTokens();
      if (!hasToken) {
        state = state.copyWith(isLoading: false, isAuthenticated: false);
        return;
      }
      final user = await _repo.getMe();
      state = state.copyWith(
        isLoading: false,
        isAuthenticated: true,
        user: user,
      );
    } catch (e) {
      // Стираем вход только если сервер отверг сессию. Сервер недоступен
      // (нет интернета, бэкенд не запущен) — не повод заново просить SMS:
      // токены остаются, SplashPage покажет ошибку с «Повторить».
      final rejected = e is DioException && e.response?.statusCode == 401;
      if (rejected) await _storage.clearTokens();
      state = state.copyWith(
        isLoading: false,
        isAuthenticated: false,
        error: rejected ? null : _message(e),
      );
    }
  }

  Future<void> sendCode(String phone) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      await _repo.sendCode(phone);
      state = state.copyWith(isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _message(e));
    }
  }

  Future<bool> verify(String phone, String code) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final result = await _repo.verify(phone, code);
      state = state.copyWith(
        isLoading: false,
        isAuthenticated: true,
        user: result.user,
      );
      return result.isNewUser;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _message(e));
      return false;
    }
  }

  Future<bool> googleSignIn(String idToken) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final result = await _repo.googleSignIn(idToken);
      state = state.copyWith(
        isLoading: false,
        isAuthenticated: true,
        user: result.user,
      );
      return result.isNewUser;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _message(e));
      return false;
    }
  }

  /// Сохраняет роль. Возвращает пользователя с ролью или null при ошибке
  /// (текст — в state.error).
  Future<User?> setRole(String role) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final user = await _repo.setRole(role);
      state = state.copyWith(isLoading: false, user: user);
      return user;
    } catch (e) {
      if (!_isRoleAlreadySet(e)) {
        state = state.copyWith(isLoading: false, error: _message(e));
        return null;
      }
    }
    // Роль уже сохранена (повторное нажатие, другая вкладка) — это не ошибка:
    // берём её с сервера и перевыпускаем токен, в котором она записана.
    try {
      final refreshToken = await _storage.getRefreshToken();
      if (refreshToken != null) await _repo.refresh(refreshToken);
      final user = await _repo.getMe();
      state = state.copyWith(isLoading: false, user: user);
      return user;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _message(e));
      return null;
    }
  }

  /// Текст последней ошибки — для экранов, которые уже не следят за state.
  String? get lastError => state.error;

  static bool _isRoleAlreadySet(Object e) {
    if (e is! DioException) return false;
    final data = e.response?.data;
    if (data is! Map) return false;
    final error = data['error'];
    return error is Map && error['code'] == 'ROLE_ALREADY_SET';
  }

  Future<void> updateProfile({String? name, String? language}) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final user = await _repo.updateMe(name: name, language: language);
      state = state.copyWith(isLoading: false, user: user);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _message(e));
    }
  }

  /// Аватар: фото (только взрослые) или животное. Возвращают текст ошибки
  /// или null при успехе.
  Future<String?> uploadAvatar({
    String? path,
    Uint8List? bytes,
    required String name,
  }) => _avatar(_repo.uploadAvatar(path: path, bytes: bytes, name: name));

  Future<String?> setAvatarPreset(String preset) =>
      _avatar(_repo.setAvatarPreset(preset));

  Future<String?> deleteAvatar() => _avatar(_repo.deleteAvatar());

  Future<String?> _avatar(Future<User> request) async {
    try {
      final user = await request;
      state = state.copyWith(user: user);
      return null;
    } catch (e) {
      return _message(e);
    }
  }

  /// Меняет имя в профиле. Возвращает текст ошибки или null при успехе.
  Future<String?> rename(String name) async {
    try {
      final user = await _repo.updateMe(name: name);
      state = state.copyWith(user: user);
      return null;
    } catch (e) {
      return _message(e);
    }
  }

  /// Меняет @username репетитора ("" — убрать). Текст ошибки или null.
  Future<String?> setUsername(String username) async {
    try {
      final user = await _repo.updateMe(username: username);
      state = state.copyWith(user: user);
      return null;
    } catch (e) {
      return _message(e);
    }
  }

  /// Отмечает на сервере ученика младше 13 (по возрастному экрану на этом
  /// устройстве): его заявки в группы будет разрешать родитель. Снять отметку
  /// сервер не даёт. Ошибка не мешает — повторим при следующем входе.
  Future<void> markChild() async {
    final user = state.user;
    if (user == null || !user.isStudent || user.isChild) return;
    try {
      final updated = await _repo.updateMe(isChild: true);
      if (state.user?.id == updated.id) state = state.copyWith(user: updated);
    } catch (e) {
      debugPrint('mark child: $e');
    }
  }

  /// Тихо сохраняет язык интерфейса в профиле: по нему сервер выбирает язык
  /// push. Ошибка не мешает работе — повторим при следующем входе или смене языка.
  Future<void> syncLanguage(String language) async {
    final user = state.user;
    if (user == null || user.language == language) return;
    try {
      final updated = await _repo.updateMe(language: language);
      if (state.user?.id == updated.id) state = state.copyWith(user: updated);
    } catch (e) {
      debugPrint('language sync: $e');
    }
  }

  Future<void> logout() async {
    final refreshToken = await _storage.getRefreshToken();
    if (refreshToken != null) {
      try {
        await _repo.logout(refreshToken);
      } catch (_) {
        // Сервер недоступен — refresh-токен истечёт сам; выходим локально.
      }
    }
    // _repo.logout очищает токены только при успехе запроса — чистим всегда.
    await _storage.clearTokens();
    state = const AuthState();
  }

  /// Удаляет аккаунт. Возвращает текст ошибки или null при успехе
  /// (тогда пользователь разлогинен).
  Future<String?> deleteAccount() async {
    try {
      await _repo.deleteAccount();
      state = const AuthState();
      return null;
    } catch (e) {
      return _message(e);
    }
  }

  /// Сервер отверг сессию посреди работы (токены уже стёрты) — на вход.
  void sessionExpired() {
    if (state.isAuthenticated) state = const AuthState();
  }

  String _message(Object e) => userErrorMessage(e);
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final notifier = AuthNotifier(
    ref.watch(authRepositoryProvider),
    ref.watch(secureStorageProvider),
  );
  ref.listen(sessionExpiredProvider, (_, __) => notifier.sessionExpired());
  return notifier;
});

/// Веб: если в другой вкладке вошли в другой аккаунт или вышли — эта вкладка
/// перезагружается. Подключён в App.
final sessionTabsSyncProvider = Provider<void>((ref) {
  onSessionUserChangedElsewhere((userId) {
    if (userId != ref.read(authProvider).user?.id) reloadApp();
  });
  ref.listen(authProvider.select((s) => s.user?.id), (previous, id) {
    if (previous != id) publishSessionUser(id);
  });
});

/// Держит язык профиля на сервере равным языку интерфейса: после входа и при
/// смене языка. Подключён в App — одно место вместо вызовов по экранам.
final languageSyncProvider = Provider<void>((ref) {
  void sync() => ref
      .read(authProvider.notifier)
      .syncLanguage(ref.read(localeProvider).languageCode);
  ref.listen(localeProvider, (_, __) => sync());
  ref.listen(authProvider.select((s) => s.user?.id), (_, id) {
    if (id != null) sync();
  });
});
