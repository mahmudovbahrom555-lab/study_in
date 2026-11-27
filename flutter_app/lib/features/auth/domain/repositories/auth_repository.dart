import 'dart:typed_data';

import '../entities/user.dart';

abstract class AuthRepository {
  Future<void> sendCode(String phone);

  Future<({String accessToken, String refreshToken, User user, bool isNewUser})>
  verify(String phone, String code, {String deviceInfo});

  Future<({String accessToken, String refreshToken, User user, bool isNewUser})>
  googleSignIn(String idToken, {String deviceInfo});

  Future<({String accessToken, String refreshToken})> refresh(
    String refreshToken,
  );

  Future<void> logout(String refreshToken);

  Future<User> setRole(String role);

  Future<User> getMe();

  Future<User> updateMe({
    String? name,
    String? language,
    String? username,
    bool? isChild,
  });

  /// Фото профиля: путь (телефон) или содержимое (веб).
  Future<User> uploadAvatar({
    String? path,
    Uint8List? bytes,
    required String name,
  });

  Future<User> setAvatarPreset(String preset);

  Future<User> deleteAvatar();

  /// Удаляет аккаунт со всеми данными и стирает вход на устройстве.
  Future<void> deleteAccount();

  Future<void> addDeviceToken(String token, String platform);
}
