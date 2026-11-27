import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../domain/entities/send_code_result.dart';
import '../after_sign_in.dart';
import '../providers/auth_provider.dart';
import '../widgets/code_input.dart';
import '../../../../core/localization/l10n.dart';

/// Что передаёт экран телефона экрану кода.
class VerifyArgs {
  const VerifyArgs({required this.phone, required this.sent});

  final String phone;
  final SendCodeResult sent;
}

class VerifyPage extends ConsumerStatefulWidget {
  const VerifyPage({super.key, required this.phone, required this.sent});

  final String phone;
  final SendCodeResult sent;

  @override
  ConsumerState<VerifyPage> createState() => _VerifyPageState();
}

class _VerifyPageState extends ConsumerState<VerifyPage> {
  String _code = '';
  late SendCodeResult _sent = widget.sent;

  Future<void> _openBot() async {
    final url = _sent.botUrl;
    if (url == null) return;
    await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
  }

  Future<void> _resend() async {
    final messenger = ScaffoldMessenger.of(context);
    final resent = context.l10n.codeResent;
    final sent = await ref.read(authProvider.notifier).sendCode(widget.phone);
    if (!mounted || sent == null) return;
    setState(() => _sent = sent);
    messenger.showSnackBar(SnackBar(content: Text(resent)));
  }

  Future<void> _verify() async {
    if (_code.length != 6) return;

    final isNewUser = await ref
        .read(authProvider.notifier)
        .verify(widget.phone, _code);
    if (!mounted) return;

    final error = ref.read(authProvider).error;
    if (error != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error), backgroundColor: Colors.red),
      );
      return;
    }

    goAfterSignIn(context, ref.read(authProvider).user, isNewUser: isNewUser);
  }

  @override
  Widget build(BuildContext context) {
    final isLoading = ref.watch(authProvider).isLoading;

    return Scaffold(
      appBar: AppBar(backgroundColor: Colors.transparent, elevation: 0),
      body: SafeArea(
        // Прокрутка — чтобы клавиатура не обрезала экран на маленьких телефонах.
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 16),
              Text(
                context.l10n.enterCodeTitle,
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                _sent.viaTelegram
                    ? context.l10n.telegramSentTo(widget.phone)
                    : context.l10n.smsSentTo(widget.phone),
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(color: Colors.grey),
              ),
              if (_sent.needsBot) ...[
                const SizedBox(height: 16),
                _OpenBotCard(onOpen: _openBot),
              ],
              const SizedBox(height: 40),
              CodeInput(
                onCompleted: (code) {
                  setState(() => _code = code);
                  _verify();
                },
              ),
              const SizedBox(height: 24),
              if (isLoading) const Center(child: CircularProgressIndicator()),
              const SizedBox(height: 24),
              TextButton(
                onPressed: isLoading ? null : _resend,
                child: Text(context.l10n.resendCode),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Номер ещё не подтверждён в боте: открыть бота и нажать «Отправить номер» —
/// бот сразу пришлёт код. Это нужно один раз.
class _OpenBotCard extends StatelessWidget {
  const _OpenBotCard({required this.onOpen});

  final VoidCallback onOpen;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(context.l10n.telegramOpenBotHint),
            const SizedBox(height: 12),
            FilledButton.icon(
              icon: const Icon(Icons.send),
              label: Text(context.l10n.telegramOpenBot),
              onPressed: onOpen,
            ),
          ],
        ),
      ),
    );
  }
}
