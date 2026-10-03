import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/storage/secure_storage.dart';
import '../../data/auth_api.dart';
import '../../data/auth_repository_impl.dart';
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

  Future<void> setRole(String role) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final user = await _repo.setRole(role);
      state = state.copyWith(isLoading: false, user: user);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: _message(e));
    }
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

  String _message(Object e) => userErrorMessage(e);
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(
    ref.watch(authRepositoryProvider),
    ref.watch(secureStorageProvider),
  );
});
