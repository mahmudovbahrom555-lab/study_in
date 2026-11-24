import 'package:flutter/material.dart';
import 'package:google_sign_in_web/web_only.dart' as web;

import '../../data/google_auth.dart';

/// На вебе нажатие обрабатывает Google; токен приходит в
/// GoogleAuth.webIdTokens, поэтому [onPressed] и [label] не используются.
class GoogleButton extends StatelessWidget {
  const GoogleButton({super.key, required this.label, this.onPressed});

  final String label;
  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<void>(
      future: GoogleAuth.instance.ensureInitialized(),
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done ||
            snapshot.hasError) {
          return const SizedBox(height: 44);
        }
        return Center(child: web.renderButton());
      },
    );
  }
}
