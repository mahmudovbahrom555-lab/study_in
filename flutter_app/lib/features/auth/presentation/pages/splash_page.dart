import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';

/// Стартовый экран. На Этапе 0 проверяет соединение с бэкендом
/// и показывает результат — чтобы убедиться что вся инфраструктура работает.
///
/// На Этапе 1 заменится логикой проверки JWT и переходом на нужный экран.
class SplashPage extends ConsumerStatefulWidget {
  const SplashPage({super.key});

  @override
  ConsumerState<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends ConsumerState<SplashPage> {
  String _status = 'Подключение...';
  bool _isError = false;
  bool _isLoading = true;
  late final Dio _dio;

  @override
  void initState() {
    super.initState();
    _dio = createDio();
    _checkBackend();
  }

  @override
  void dispose() {
    _dio.close();
    super.dispose();
  }

  Future<void> _checkBackend() async {
    setState(() => _isLoading = true);
    try {
      final response = await _dio.get<Map<String, dynamic>>('/health');

      if (!mounted) return;

      setState(() {
        _status = '✓ Backend работает\nстатус: ${response.data?['data']?['status']}';
        _isError = false;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _status = 'Не удалось подключиться:\n$e';
        _isError = true;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const FlutterLogo(size: 80),
              const SizedBox(height: 32),
              const Text(
                'RepetApp',
                style: TextStyle(
                  fontSize: 32,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              const Text(
                'Этап 0: проверка инфраструктуры',
                style: TextStyle(fontSize: 14, color: Colors.grey),
              ),
              const SizedBox(height: 48),
              if (_isLoading) const CircularProgressIndicator(),
              const SizedBox(height: 24),
              Text(
                _status,
                textAlign: TextAlign.center,
                style: TextStyle(
                  color: _isError ? Colors.red : Colors.green,
                  fontSize: 14,
                ),
              ),
              const SizedBox(height: 24),
              if (_isError)
                ElevatedButton(
                  onPressed: _checkBackend,
                  child: const Text('Повторить'),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
