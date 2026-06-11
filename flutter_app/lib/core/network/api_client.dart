import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../config/app_config.dart';
import '../storage/secure_storage.dart';
import 'interceptors/auth_interceptor.dart';
import 'interceptors/logging_interceptor.dart';

final dioProvider = Provider<Dio>((ref) => createDio(ref));

Dio createDio([Ref? ref]) {
  final dio = Dio(
    BaseOptions(
      baseUrl: AppConfig.apiUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
      sendTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      validateStatus: (status) => status != null && status >= 200 && status < 300,
    ),
  );

  if (AppConfig.isDev) {
    dio.interceptors.add(LoggingInterceptor());
  }

  if (ref != null) {
    final storage = ref.read(secureStorageProvider);
    dio.interceptors.add(AuthInterceptor(storage, dio));
  }

  return dio;
}
