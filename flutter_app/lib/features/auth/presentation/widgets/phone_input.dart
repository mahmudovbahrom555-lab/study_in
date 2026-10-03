import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../core/localization/l10n.dart';

class PhoneInput extends StatelessWidget {
  const PhoneInput({
    super.key,
    required this.controller,
    required this.onChanged,
    this.onSubmitted,
  });

  final TextEditingController controller;
  final ValueChanged<String> onChanged;
  final ValueChanged<String>? onSubmitted;

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      keyboardType: TextInputType.phone,
      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
      maxLength: 13,
      decoration: InputDecoration(
        prefixText: '+998 ',
        labelText: context.l10n.phoneNumber,
        hintText: '90 123 45 67',
        border: const OutlineInputBorder(),
        counterText: '',
      ),
      onChanged: (v) => onChanged('+998$v'),
      onFieldSubmitted: onSubmitted,
    );
  }
}
