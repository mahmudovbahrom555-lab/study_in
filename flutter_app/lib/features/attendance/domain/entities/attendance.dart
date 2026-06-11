import 'package:flutter/foundation.dart';

enum AttendanceStatus {
  present,
  absent,
  late,
  excused;

  static AttendanceStatus fromJson(String v) => switch (v) {
        'absent' => absent,
        'late' => late,
        'excused' => excused,
        _ => present,
      };

  String toJson() => name;

  String get label => switch (this) {
        present => 'Присутствует',
        absent => 'Отсутствует',
        late => 'Опоздал',
        excused => 'Уважительная',
      };
}

@immutable
class Attendance {
  const Attendance({
    required this.id,
    required this.groupId,
    required this.studentId,
    required this.teacherId,
    required this.lessonDate,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
    this.note,
  });

  final String id;
  final String groupId;
  final String studentId;
  final String teacherId;
  final String lessonDate; // "2024-09-01"
  final AttendanceStatus status;
  final String? note;
  final DateTime createdAt;
  final DateTime updatedAt;
}
