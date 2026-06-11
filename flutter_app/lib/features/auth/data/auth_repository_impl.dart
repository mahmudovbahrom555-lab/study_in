import '../../../core/storage/secure_storage.dart';
import '../domain/entities/user.dart';
import '../domain/repositories/auth_repository.dart';
import 'auth_api.dart';

class AuthRepositoryImpl implements AuthRepository {
  AuthRepositoryImpl(this._api, this._storage);

  final AuthApi _api;
  final SecureStorage _storage;

  @override
  Future<void> sendCode(String phone) => _api.sendCode(phone);

  @override
  Future<({String accessToken, String refreshToken, User user, bool isNewUser})> verify(
    String phone,
    String code, {
    String deviceInfo = '',
  }) async {
    final dto = await _api.verify(phone, code, deviceInfo: deviceInfo);
    await _storage.saveTokens(
      accessToken: dto.accessToken,
      refreshToken: dto.refreshToken,
    );
    return (
      accessToken: dto.accessToken,
      refreshToken: dto.refreshToken,
      user: dto.user.toEntity(),
      isNewUser: dto.isNewUser,
    );
  }

  @override
  Future<({String accessToken, String refreshToken})> refresh(
    String refreshToken,
  ) async {
    final tokens = await _api.refresh(refreshToken);
    await _storage.saveTokens(
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
    );
    return tokens;
  }

  @override
  Future<void> logout(String refreshToken) async {
    await _api.logout(refreshToken);
    await _storage.clearTokens();
  }

  @override
  Future<User> setRole(String role) async {
    final dto = await _api.setRole(role);
    return dto.toEntity();
  }

  @override
  Future<User> getMe() async {
    final dto = await _api.getMe();
    return dto.toEntity();
  }

  @override
  Future<User> updateMe({String? name, String? language}) async {
    final dto = await _api.updateMe(name: name, language: language);
    return dto.toEntity();
  }

  @override
  Future<void> addDeviceToken(String token, String platform) =>
      _api.addDeviceToken(token, platform);
}
