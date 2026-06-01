import 'package:flutter/material.dart';

/// Цветовая палитра приложения.
/// Минималистичная, с акцентом на синий (доверие, обучение).
class AppColors {
  AppColors._();

  // Основные
  static const Color primary = Color(0xFF2563EB); // Синий
  static const Color primaryDark = Color(0xFF1E40AF);
  static const Color primaryLight = Color(0xFF60A5FA);

  // Семантические
  static const Color success = Color(0xFF10B981); // Зелёный
  static const Color warning = Color(0xFFF59E0B); // Оранжевый
  static const Color error = Color(0xFFEF4444); // Красный
  static const Color info = Color(0xFF3B82F6);

  // Нейтральные
  static const Color background = Color(0xFFF7F9FC);
  static const Color surface = Colors.white;
  static const Color surfaceDark = Color(0xFF1F2937);
  static const Color border = Color(0xFFE5E7EB);
  static const Color divider = Color(0xFFF3F4F6);

  // Текст
  static const Color textPrimary = Color(0xFF111827);
  static const Color textSecondary = Color(0xFF6B7280);
  static const Color textHint = Color(0xFF9CA3AF);
  static const Color textOnPrimary = Colors.white;

  // Карточки ленты (по типу)
  static const Color announcementBg = Color(0xFFF3F4F6);
  static const Color assignmentBg = Color(0xFFEFF6FF);
  static const Color testBg = Color(0xFFECFDF5);
  static const Color resultBg = Color(0xFFFAF5FF);
}
