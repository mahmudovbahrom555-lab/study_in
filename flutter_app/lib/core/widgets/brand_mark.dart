import 'package:flutter/material.dart';

import '../localization/l10n.dart';
import '../theme/colors.dart';

/// Знак «Росток» — те же кривые, что в brand/logo/symbol.svg (сетка 64×64).
class BrandMark extends StatelessWidget {
  const BrandMark({super.key, this.size = 24, this.color = AppColors.primary});

  final double size;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return CustomPaint(size: Size.square(size), painter: _SproutPainter(color));
  }
}

class _SproutPainter extends CustomPainter {
  const _SproutPainter(this.color);

  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.scale(size.width / 64);
    final paint = Paint()..color = color;
    canvas.drawPath(
      Path()
        ..moveTo(30, 56)
        ..cubicTo(30, 34, 40, 14, 58, 8)
        ..cubicTo(60, 30, 50, 50, 30, 56)
        ..close(),
      paint,
    );
    canvas.drawPath(
      Path()
        ..moveTo(26, 56)
        ..cubicTo(10, 54, 4, 40, 6, 26)
        ..cubicTo(20, 30, 28, 42, 26, 56)
        ..close(),
      paint,
    );
  }

  @override
  bool shouldRepaint(_SproutPainter old) => old.color != color;
}

/// Логотип для шапки: знак + «MaMu Learn».
class BrandTitle extends StatelessWidget {
  const BrandTitle({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final base = theme.textTheme.titleLarge;
    return Semantics(
      label: context.l10n.appName,
      header: true,
      excludeSemantics: true,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const BrandMark(size: 28),
          const SizedBox(width: 8),
          Text.rich(
            TextSpan(
              children: [
                TextSpan(text: 'MaMu ', style: base?.copyWith(fontWeight: FontWeight.w700)),
                TextSpan(
                  text: 'Learn',
                  style: base?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: theme.colorScheme.primary,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
