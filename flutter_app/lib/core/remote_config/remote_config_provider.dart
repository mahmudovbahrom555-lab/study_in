import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:package_info_plus/package_info_plus.dart';

import '../localization/locale_provider.dart' show settingsBoxName;
import '../network/api_client.dart';
import 'remote_config.dart';

/// Установленное приложение (версия, id пакета). Подставляется в main().
final appInfoProvider = Provider<PackageInfo>(
  (ref) => throw UnimplementedError('appInfoProvider задаётся в main()'),
);

/// Последняя полученная с сервера конфигурация; до первого ответа — из кэша.
class RemoteConfigController extends StateNotifier<RemoteConfig> {
  RemoteConfigController(this._dio, this._box) : super(_cached(_box));

  static const _key = 'remote_config';
  final Dio _dio;
  final Box<dynamic> _box;

  static RemoteConfig _cached(Box<dynamic> box) {
    final raw = box.get(_key) as String?;
    if (raw == null) return const RemoteConfig();
    try {
      return RemoteConfig.fromJson(jsonDecode(raw) as Map<String, dynamic>);
    } catch (_) {
      return const RemoteConfig();
    }
  }

  /// Без сети остаётся конфигурация из кэша — запуск не блокируем.
  Future<void> refresh() async {
    try {
      final res = await _dio.get<Map<String, dynamic>>('/app-config');
      final config = RemoteConfig.fromJson(res.data!['data'] as Map<String, dynamic>);
      state = config;
      await _box.put(_key, jsonEncode(config.toJson()));
    } catch (e) {
      debugPrint('app-config: $e');
    }
  }
}

final remoteConfigProvider =
    StateNotifierProvider<RemoteConfigController, RemoteConfig>(
  (ref) => RemoteConfigController(
    ref.watch(dioProvider),
    Hive.box<dynamic>(settingsBoxName),
  ),
);
