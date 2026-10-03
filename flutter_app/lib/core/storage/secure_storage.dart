import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _kAccessToken = 'access_token';
const _kRefreshToken = 'refresh_token';

class SecureStorage {
  SecureStorage(this._storage);

  final FlutterSecureStorage _storage;

  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    // Последовательно, не через Future.wait: на web первая запись генерирует
    // ключ шифрования, и параллельные записи создают разные ключи —
    // одно из значений потом не расшифровывается (OperationError).
    await _storage.write(key: _kAccessToken, value: accessToken);
    await _storage.write(key: _kRefreshToken, value: refreshToken);
  }

  Future<String?> getAccessToken() => _read(_kAccessToken);
  Future<String?> getRefreshToken() => _read(_kRefreshToken);

  Future<void> clearTokens() async {
    await _storage.delete(key: _kAccessToken);
    await _storage.delete(key: _kRefreshToken);
  }

  Future<bool> hasTokens() async {
    final token = await _read(_kAccessToken);
    return token != null && token.isNotEmpty;
  }

  /// Повреждённое хранилище (не расшифровывается) — сбрасываем его,
  /// пользователь просто залогинится заново.
  Future<String?> _read(String key) async {
    try {
      return await _storage.read(key: key);
    } catch (_) {
      await _storage.deleteAll();
      return null;
    }
  }
}

final secureStorageProvider = Provider<SecureStorage>((ref) {
  return SecureStorage(const FlutterSecureStorage());
});
