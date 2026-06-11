import 'package:dio/dio.dart';

import 'models/user_dto.dart';

class AuthApi {
  AuthApi(this._dio);

  final Dio _dio;

  Future<void> sendCode(String phone) async {
    await _dio.post<void>('/auth/send-code', data: {'phone': phone});
  }

  Future<AuthResponseDto> verify(
    String phone,
    String code, {
    String deviceInfo = '',
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/auth/verify',
      data: {'phone': phone, 'code': code, 'device_info': deviceInfo},
    );
    return AuthResponseDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<({String accessToken, String refreshToken})> refresh(
    String refreshToken,
  ) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/auth/refresh',
      data: {'refresh_token': refreshToken},
    );
    final data = resp.data!['data'] as Map<String, dynamic>;
    return (
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
  }

  Future<void> logout(String refreshToken) async {
    await _dio.post<void>('/auth/logout', data: {'refresh_token': refreshToken});
  }

  Future<UserDto> setRole(String role) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/auth/role',
      data: {'role': role},
    );
    return UserDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<UserDto> getMe() async {
    final resp = await _dio.get<Map<String, dynamic>>('/me');
    return UserDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<UserDto> updateMe({String? name, String? language}) async {
    final body = <String, dynamic>{
      if (name != null) 'name': name,
      if (language != null) 'language': language,
    };
    final resp = await _dio.patch<Map<String, dynamic>>('/me', data: body);
    return UserDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> addDeviceToken(String token, String platform) async {
    await _dio.post<void>(
      '/me/device-token',
      data: {'token': token, 'platform': platform},
    );
  }
}
