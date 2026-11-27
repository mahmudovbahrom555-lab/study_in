import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/config/app_config.dart';
import '../../../../core/localization/language_switcher.dart';
import '../../../../core/network/error_message.dart';
import '../../../../core/router/routes.dart';
import '../../data/birth_year.dart';
import '../../data/google_auth.dart';
import '../after_sign_in.dart';
import '../providers/auth_provider.dart';
import '../widgets/age_gate.dart';
import '../widgets/google_button.dart';
import '../widgets/phone_input.dart';
import 'verify_page.dart';
import '../../../../core/localization/l10n.dart';

/// Согласие с политикой: вход = регистрация, поэтому согласие стоит здесь.
/// Его версию приложение отправляет при входе (AppConfig.privacyVersion).
class _ConsentNote extends StatefulWidget {
  const _ConsentNote();

  @override
  State<_ConsentNote> createState() => _ConsentNoteState();
}

class _ConsentNoteState extends State<_ConsentNote> {
  final _tap = TapGestureRecognizer()
    ..onTap = () => launchUrl(
      Uri.parse(AppConfig.privacyUrl),
      mode: LaunchMode.externalApplication,
    );

  @override
  void dispose() {
    _tap.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Text.rich(
      TextSpan(
        style: theme.textTheme.bodySmall?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
        children: [
          TextSpan(text: context.l10n.loginConsent),
          TextSpan(
            text: context.l10n.privacyPolicyLink,
            style: TextStyle(
              color: theme.colorScheme.primary,
              decoration: TextDecoration.underline,
            ),
            recognizer: _tap,
          ),
        ],
      ),
      textAlign: TextAlign.center,
    );
  }
}

class _OrDivider extends StatelessWidget {
  const _OrDivider({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        const Expanded(child: Divider()),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: Text(label, style: const TextStyle(color: Colors.grey)),
        ),
        const Expanded(child: Divider()),
      ],
    );
  }
}

class PhonePage extends ConsumerStatefulWidget {
  const PhonePage({super.key});

  @override
  ConsumerState<PhonePage> createState() => _PhonePageState();
}

class _PhonePageState extends ConsumerState<PhonePage> {
  final _controller = TextEditingController();
  String _phone = '';
  StreamSubscription<String>? _webGoogle;

  @override
  void initState() {
    super.initState();
    if (kIsWeb && GoogleAuth.instance.isConfigured) {
      _webGoogle = GoogleAuth.instance.webIdTokens.listen(_googleSignIn);
    }
  }

  @override
  void dispose() {
    _webGoogle?.cancel();
    _controller.dispose();
    super.dispose();
  }

  Future<void> _startGoogle() async {
    try {
      final idToken = await GoogleAuth.instance.signIn();
      if (idToken != null) await _googleSignIn(idToken);
    } catch (e) {
      _showError(userErrorMessage(e));
    }
  }

  Future<void> _googleSignIn(String idToken) async {
    final notifier = ref.read(authProvider.notifier);
    final isNewUser = await notifier.googleSignIn(idToken);
    if (!mounted) return;
    final state = ref.read(authProvider);
    if (state.error != null) {
      _showError(state.error!);
      return;
    }
    goAfterSignIn(context, state.user, isNewUser: isNewUser);
  }

  void _showError(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: Colors.red),
    );
  }

  Future<void> _sendCode() async {
    final phone = _phone.trim();
    if (phone.length < 9) return;

    final sent = await ref.read(authProvider.notifier).sendCode(phone);
    if (!mounted) return;

    final error = ref.read(authProvider).error;
    if (error != null || sent == null) {
      _showError(error ?? context.l10n.errGeneric);
      return;
    }
    context.push(
      Routes.verify,
      extra: VerifyArgs(phone: phone, sent: sent),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isLoading = ref.watch(authProvider).isLoading;
    final birthYear = ref.watch(birthYearProvider);
    if (birthYear == null) {
      return Scaffold(
        body: SafeArea(
          child: AgeGate(
            onSelected: (y) => ref.read(birthYearProvider.notifier).select(y),
          ),
        ),
      );
    }
    // Политика Google Play для семей: детям вход через Google не предлагаем.
    final showGoogle =
        GoogleAuth.instance.isConfigured && !isChildBirthYear(birthYear);

    // Прокрутка: с открытой клавиатурой или крупным системным шрифтом экран
    // не помещается, и без неё низ (вход через Google, согласие) обрезается.
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Язык можно сменить до входа: автоопределение иногда ошибается.
              const Align(
                alignment: Alignment.centerRight,
                child: LanguageSwitcher(),
              ),
              const SizedBox(height: 16),
              Text(
                context.l10n.loginTitle,
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                context.l10n.loginSubtitle,
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(color: Colors.grey),
              ),
              const SizedBox(height: 40),
              PhoneInput(
                controller: _controller,
                onChanged: (v) => setState(() => _phone = v),
                onSubmitted: (_) => _sendCode(),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton(
                  onPressed: isLoading || _phone.length < 9 ? null : _sendCode,
                  child: isLoading
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Text(context.l10n.getCode),
                ),
              ),
              if (showGoogle) ...[
                const SizedBox(height: 16),
                _OrDivider(label: context.l10n.orDivider),
                const SizedBox(height: 16),
                GoogleButton(
                  label: context.l10n.signInWithGoogle,
                  onPressed: isLoading ? null : _startGoogle,
                ),
              ],
              const SizedBox(height: 16),
              const _ConsentNote(),
            ],
          ),
        ),
      ),
    );
  }
}
