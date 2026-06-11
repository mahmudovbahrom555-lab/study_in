import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

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
      decoration: const InputDecoration(
        prefixText: '+998 ',
        labelText: 'Номер телефона',
        hintText: '90 123 45 67',
        border: OutlineInputBorder(),
        counterText: '',
      ),
      onChanged: (v) => onChanged('+998$v'),
      onFieldSubmitted: onSubmitted,
    );
  }
}
