import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../providers/auth_provider.dart';
import '../widgets/code_input.dart';

class VerifyPage extends ConsumerStatefulWidget {
  const VerifyPage({super.key, required this.phone});

  final String phone;

  @override
  ConsumerState<VerifyPage> createState() => _VerifyPageState();
}

class _VerifyPageState extends ConsumerState<VerifyPage> {
  String _code = '';

  Future<void> _verify() async {
    if (_code.length != 6) return;

    final isNewUser = await ref.read(authProvider.notifier).verify(widget.phone, _code);
    if (!mounted) return;

    final error = ref.read(authProvider).error;
    if (error != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error), backgroundColor: Colors.red),
      );
      return;
    }

    if (isNewUser) {
      context.go(Routes.roleSelect);
    } else {
      final user = ref.read(authProvider).user;
      if (user != null && user.name.isEmpty) {
        context.go(Routes.profileSetup);
      } else {
        context.go(Routes.home);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final isLoading = ref.watch(authProvider).isLoading;

    return Scaffold(
      appBar: AppBar(backgroundColor: Colors.transparent, elevation: 0),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 16),
              Text(
                'Введите код',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              const SizedBox(height: 8),
              Text(
                'Отправили SMS на ${widget.phone}',
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: Colors.grey),
              ),
              const SizedBox(height: 40),
              CodeInput(
                onCompleted: (code) {
                  setState(() => _code = code);
                  _verify();
                },
              ),
              const SizedBox(height: 24),
              if (isLoading) const Center(child: CircularProgressIndicator()),
              const Spacer(),
              TextButton(
                onPressed: () {
                  ref.read(authProvider.notifier).sendCode(widget.phone);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Код отправлен повторно')),
                  );
                },
                child: const Text('Отправить код повторно'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
