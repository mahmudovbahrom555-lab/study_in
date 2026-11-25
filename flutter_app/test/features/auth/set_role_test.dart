import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/storage/secure_storage.dart';
import 'package:mamu_learn/features/auth/domain/entities/user.dart';
import 'package:mamu_learn/features/auth/domain/repositories/auth_repository.dart';
import 'package:mamu_learn/features/auth/presentation/providers/auth_provider.dart';

const _withRole = User(
  id: 'u1',
  phone: '',
  email: 'a@gmail.com',
  name: 'Ali',
  role: 'teacher',
  language: 'uz',
  isActive: true,
);

DioException _apiError(int status, String code) {
  final options = RequestOptions(path: '/auth/role');
  return DioException(
    requestOptions: options,
    response: Response(
      requestOptions: options,
      statusCode: status,
      data: {
        'error': {'code': code, 'message': code},
      },
    ),
  );
}

class _Repo extends Fake implements AuthRepository {
  _Repo(this.setRoleError);

  final DioException setRoleError;
  int refreshed = 0;

  @override
  Future<User> setRole(String role) async => throw setRoleError;

  @override
  Future<({String accessToken, String refreshToken})> refresh(
    String refreshToken,
  ) async {
    refreshed++;
    return (accessToken: 'a2', refreshToken: 'r2');
  }

  @override
  Future<User> getMe() async => _withRole;
}

class _Storage extends Fake implements SecureStorage {
  @override
  Future<String?> getRefreshToken() async => 'r1';
}

void main() {
  test('роль уже сохранена — не ошибка: роль с сервера, токен обновлён', () async {
    final repo = _Repo(_apiError(409, 'ROLE_ALREADY_SET'));
    final notifier = AuthNotifier(repo, _Storage());

    final user = await notifier.setRole('teacher');

    expect(user?.role, 'teacher');
    expect(notifier.lastError, isNull);
    expect(repo.refreshed, 1);
  });

  test('другая ошибка — показывается, пользователь не возвращается', () async {
    final repo = _Repo(_apiError(500, 'INTERNAL'));
    final notifier = AuthNotifier(repo, _Storage());

    final user = await notifier.setRole('teacher');

    expect(user, isNull);
    expect(notifier.lastError, isNotNull);
    expect(repo.refreshed, 0);
  });
}
