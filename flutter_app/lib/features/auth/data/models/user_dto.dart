import '../../domain/entities/user.dart';

class UserDto {
  const UserDto({
    required this.id,
    required this.phone,
    required this.name,
    this.role,
    this.avatarUrl,
    required this.language,
    required this.isActive,
  });

  factory UserDto.fromJson(Map<String, dynamic> json) {
    return UserDto(
      id: json['id'] as String,
      phone: json['phone'] as String,
      name: (json['name'] as String?) ?? '',
      role: json['role'] as String?,
      avatarUrl: json['avatar_url'] as String?,
      language: (json['language'] as String?) ?? 'uz',
      isActive: (json['is_active'] as bool?) ?? true,
    );
  }

  final String id;
  final String phone;
  final String name;
  final String? role;
  final String? avatarUrl;
  final String language;
  final bool isActive;

  User toEntity() => User(
        id: id,
        phone: phone,
        name: name,
        role: role,
        avatarUrl: avatarUrl,
        language: language,
        isActive: isActive,
      );
}

class AuthResponseDto {
  const AuthResponseDto({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
    required this.isNewUser,
  });

  factory AuthResponseDto.fromJson(Map<String, dynamic> json) {
    return AuthResponseDto(
      accessToken: json['access_token'] as String,
      refreshToken: json['refresh_token'] as String,
      user: UserDto.fromJson(json['user'] as Map<String, dynamic>),
      isNewUser: (json['is_new_user'] as bool?) ?? false,
    );
  }

  final String accessToken;
  final String refreshToken;
  final UserDto user;
  final bool isNewUser;
}
