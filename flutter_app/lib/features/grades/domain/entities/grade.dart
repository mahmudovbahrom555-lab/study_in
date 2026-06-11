import 'package:flutter/foundation.dart';

@immutable
class Grade {
  const Grade({
    required this.id,
    required this.groupId,
    required this.studentId,
    required this.teacherId,
    required this.subject,
    required this.value,
    required this.maxValue,
    required this.gradedAt,
    required this.createdAt,
    this.comment,
  });

  final String id;
  final String groupId;
  final String studentId;
  final String teacherId;
  final String subject;
  final double value;
  final double maxValue;
  final String? comment;
  final DateTime gradedAt;
  final DateTime createdAt;

  double get percentage => maxValue > 0 ? value / maxValue * 100 : 0;
}
