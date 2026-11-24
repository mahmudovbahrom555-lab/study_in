import 'package:flutter/material.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/localization/language_switcher.dart';

/// Нейтральный возрастной экран: без подсказки о пороге и без выбранного
/// заранее года, чтобы не подталкивать ребёнка указать неверный возраст.
class AgeGate extends StatefulWidget {
  const AgeGate({super.key, required this.onSelected});

  final ValueChanged<int> onSelected;

  @override
  State<AgeGate> createState() => _AgeGateState();
}

class _AgeGateState extends State<AgeGate> {
  int? _year;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final thisYear = DateTime.now().year;
    final years = [for (var y = thisYear; y >= thisYear - 100; y--) y];

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Align(
            alignment: Alignment.centerRight,
            child: LanguageSwitcher(),
          ),
          const SizedBox(height: 16),
          Text(
            context.l10n.ageGateTitle,
            style: theme.textTheme.headlineMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            context.l10n.ageGateHint,
            style: theme.textTheme.bodyMedium?.copyWith(color: Colors.grey),
          ),
          const SizedBox(height: 40),
          DropdownButtonFormField<int>(
            initialValue: _year,
            isExpanded: true,
            hint: Text(context.l10n.ageGateYear),
            menuMaxHeight: 360,
            items: [
              for (final y in years)
                DropdownMenuItem(value: y, child: Text('$y')),
            ],
            onChanged: (y) => setState(() => _year = y),
          ),
          const SizedBox(height: 24),
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton(
              onPressed: _year == null ? null : () => widget.onSelected(_year!),
              child: Text(context.l10n.continueAction),
            ),
          ),
        ],
      ),
    );
  }
}
