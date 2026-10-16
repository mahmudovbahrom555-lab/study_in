import 'dart:ui';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/localization/locale_provider.dart';
import 'package:mamu_learn/core/localization/locale_resolver.dart';
import 'package:mamu_learn/core/storage/secure_storage.dart';
import 'package:mamu_learn/features/auth/domain/entities/user.dart';
import 'package:mamu_learn/features/auth/domain/repositories/auth_repository.dart';
import 'package:mamu_learn/features/auth/presentation/providers/auth_provider.dart';

class _Repo extends Fake implements AuthRepository {
  User user = const User(id: 'u1', phone: '+998', name: 'A', role: 'teacher', language: 'uz', isActive: true);
  final sent = <String>[];

  @override
  Future<User> getMe() async => user;

  @override
  Future<User> updateMe({String? name, String? language}) async {
    sent.add(language!);
    return user = user.copyWith(language: language);
  }
}

class _Storage extends Fake implements SecureStorage {
  @override
  Future<bool> hasTokens() async => true;
}

class _MemoryLanguageStore implements LanguageStore {
  @override
  String? read() => 'ru';

  @override
  Future<void> write(String languageCode) async {}
}

void main() {
  test('язык профиля догоняет язык интерфейса после входа и при смене', () async {
    final repo = _Repo();
    final container = ProviderContainer(
      overrides: [
        authRepositoryProvider.overrideWithValue(repo),
        secureStorageProvider.overrideWithValue(_Storage()),
        localeProvider.overrideWith((ref) => LocaleController(_MemoryLanguageStore())),
      ],
    );
    addTearDown(container.dispose);
    container.read(languageSyncProvider);

    await container.read(authProvider.notifier).checkAuth();
    await pumpEventQueue();
    expect(repo.sent, ['ru'], reason: 'на сервере был uz, в приложении ru');

    await container.read(localeProvider.notifier).select(enLocale);
    await pumpEventQueue();
    expect(repo.sent, ['ru', 'en']);
    expect(container.read(authProvider).user!.language, 'en');

    await container.read(localeProvider.notifier).select(const Locale('en'));
    await pumpEventQueue();
    expect(repo.sent, ['ru', 'en'], reason: 'язык не изменился — запроса нет');
  });
}
