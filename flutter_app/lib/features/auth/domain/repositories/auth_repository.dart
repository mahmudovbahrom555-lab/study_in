import '../entities/user.dart';

abstract class AuthRepository {
  Future<void> sendCode(String phone);

  Future<({String accessToken, String refreshToken, User user, bool isNewUser})> verify(
    String phone,
    String code, {
    String deviceInfo,
  });

  Future<({String accessToken, String refreshToken})> refresh(String refreshToken);

  Future<void> logout(String refreshToken);

  Future<User> setRole(String role);

  Future<User> getMe();

  Future<User> updateMe({String? name, String? language});

  Future<void> addDeviceToken(String token, String platform);
}
