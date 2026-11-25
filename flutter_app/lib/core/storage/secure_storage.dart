import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'web_local_storage.dart';

const _kAccessToken = 'access_token';
const _kRefreshToken = 'refresh_token';

/// Ключи в localStorage на вебе (см. web_local_storage.dart).
const _kWebPrefix = 'mamu.';

/// Токены входа. Телефон — защищённое хранилище системы (Keystore/Keychain),
/// веб — localStorage.
class SecureStorage {
  SecureStorage(this._storage);

  final FlutterSecureStorage _storage;

  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    if (kIsWeb) {
      webWrite(_kWebPrefix + _kAccessToken, accessToken);
      webWrite(_kWebPrefix + _kRefreshToken, refreshToken);
      return;
    }
    // Последовательно, не через Future.wait: параллельные записи в
    // flutter_secure_storage уже приводили к нерасшифровываемым значениям.
    await _storage.write(key: _kAccessToken, value: accessToken);
    await _storage.write(key: _kRefreshToken, value: refreshToken);
  }

  Future<String?> getAccessToken() => _read(_kAccessToken);
  Future<String?> getRefreshToken() => _read(_kRefreshToken);

  Future<void> clearTokens() async {
    if (kIsWeb) {
      webDelete(_kWebPrefix + _kAccessToken);
      webDelete(_kWebPrefix + _kRefreshToken);
    }
    await _storage.delete(key: _kAccessToken);
    await _storage.delete(key: _kRefreshToken);
  }

  Future<bool> hasTokens() async {
    final token = await _read(_kAccessToken);
    return token != null && token.isNotEmpty;
  }

  Future<String?> _read(String key) async {
    if (kIsWeb) {
      final value = webRead(_kWebPrefix + key);
      if (value != null) return value;
    }
    // Телефон; на вебе — вход, сохранённый до перехода на localStorage.
    try {
      final value = await _storage.read(key: key);
      if (kIsWeb && value != null) {
        webWrite(_kWebPrefix + key, value);
        await _storage.delete(key: key);
      }
      return value;
    } catch (_) {
      // Повреждённое хранилище (не расшифровывается) — сбрасываем его,
      // пользователь просто войдёт заново.
      await _storage.deleteAll();
      return null;
    }
  }
}

final secureStorageProvider = Provider<SecureStorage>((ref) {
  return SecureStorage(const FlutterSecureStorage());
});
