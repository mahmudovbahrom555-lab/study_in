import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/network/interceptors/auth_interceptor.dart';
import 'package:mamu_learn/core/storage/secure_storage.dart';

class _MemStorage extends SecureStorage {
  _MemStorage(this.access, this.refresh) : super(const FlutterSecureStorage());

  String? access;
  String? refresh;

  @override
  Future<void> saveTokens({required String accessToken, required String refreshToken}) async {
    access = accessToken;
    refresh = refreshToken;
  }

  @override
  Future<String?> getAccessToken() async => access;
  @override
  Future<String?> getRefreshToken() async => refresh;
  @override
  Future<void> clearTokens() async => access = refresh = null;
}

/// Сервер-заглушка: /me отвечает 200 только со свежим токеном,
/// /auth/refresh — как настроено в [refreshStatus].
class _FakeServer implements HttpClientAdapter {
  _FakeServer({required this.refreshStatus});

  final int refreshStatus;
  int refreshCalls = 0;
  int meCalls = 0;
  final _gate = Completer<void>();

  /// Держит первый /auth/refresh, чтобы второй запрос успел получить 401.
  void releaseRefresh() => _gate.complete();

  @override
  Future<ResponseBody> fetch(RequestOptions o, Stream<Uint8List>? _, Future<void>? __) async {
    ResponseBody json(int status, Object body) => ResponseBody.fromString(
          jsonEncode(body),
          status,
          headers: {Headers.contentTypeHeader: [Headers.jsonContentType]},
        );

    if (o.path == '/auth/refresh') {
      refreshCalls++;
      await _gate.future;
      if (refreshStatus != 200) return json(refreshStatus, {'error': {'code': 'UNAUTHORIZED'}});
      return json(200, {'data': {'access_token': 'fresh', 'refresh_token': 'r2'}});
    }
    meCalls++;
    final ok = o.headers['Authorization'] == 'Bearer fresh';
    return ok ? json(200, {'data': {'id': 'u1'}}) : json(401, {'error': {'code': 'UNAUTHORIZED'}});
  }

  @override
  void close({bool force = false}) {}
}

({Dio dio, _MemStorage storage, _FakeServer server, List<int> expired}) _setup(int refreshStatus) {
  final storage = _MemStorage('stale', 'r1');
  final server = _FakeServer(refreshStatus: refreshStatus);
  final expired = <int>[];
  final dio = Dio(BaseOptions(baseUrl: 'http://test'))..httpClientAdapter = server;
  final refreshDio = Dio(BaseOptions(baseUrl: 'http://test'))..httpClientAdapter = server;
  dio.interceptors.add(
    AuthInterceptor(
      storage,
      dio,
      refreshDio: refreshDio,
      onSessionExpired: () => expired.add(1),
    ),
  );
  return (dio: dio, storage: storage, server: server, expired: expired);
}

void main() {
  test('отвергнутый refresh-токен: один запрос, выход, без цикла', () async {
    final s = _setup(401);
    s.server.releaseRefresh();

    await expectLater(s.dio.get<dynamic>('/me'), throwsA(isA<DioException>()));

    expect(s.server.refreshCalls, 1, reason: 'раньше здесь были тысячи запросов');
    expect(s.storage.access, isNull);
    expect(s.storage.refresh, isNull);
    expect(s.expired, hasLength(1));
  });

  test('два одновременных 401 — одно обновление, оба запроса повторены', () async {
    final s = _setup(200);
    final a = s.dio.get<dynamic>('/me');
    final b = s.dio.get<dynamic>('/me');
    await Future<void>.delayed(const Duration(milliseconds: 50));
    s.server.releaseRefresh();

    final results = await Future.wait([a, b]);
    expect(results.map((r) => r.statusCode), [200, 200]);
    expect(s.server.refreshCalls, 1);
    expect(s.storage.refresh, 'r2');
    expect(s.expired, isEmpty);
  });

  test('обрыв сети при обновлении не выкидывает из аккаунта', () async {
    final s = _setup(503);
    s.server.releaseRefresh();

    await expectLater(s.dio.get<dynamic>('/me'), throwsA(isA<DioException>()));
    expect(s.storage.refresh, 'r1');
    expect(s.expired, isEmpty);
  });
}
