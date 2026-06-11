import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// 6-значный OTP-инпут. Вызывает [onCompleted] когда все 6 цифр введены.
class CodeInput extends StatefulWidget {
  const CodeInput({super.key, required this.onCompleted});

  final ValueChanged<String> onCompleted;

  @override
  State<CodeInput> createState() => _CodeInputState();
}

class _CodeInputState extends State<CodeInput> {
  final _controller = TextEditingController();
  final _focus = FocusNode();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _focus.requestFocus());
  }

  @override
  void dispose() {
    _controller.dispose();
    _focus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        TextField(
          controller: _controller,
          focusNode: _focus,
          keyboardType: TextInputType.number,
          inputFormatters: [
            FilteringTextInputFormatter.digitsOnly,
            LengthLimitingTextInputFormatter(6),
          ],
          textAlign: TextAlign.center,
          style: Theme.of(context)
              .textTheme
              .headlineMedium
              ?.copyWith(letterSpacing: 16),
          decoration: const InputDecoration(
            border: OutlineInputBorder(),
            hintText: '------',
            counterText: '',
          ),
          onChanged: (v) {
            if (v.length == 6) widget.onCompleted(v);
          },
        ),
      ],
    );
  }
}
