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

/// Ученик в списке занятия; [status] == null — ещё не отмечен.
@immutable
class RosterEntry {
  const RosterEntry({
    required this.studentId,
    required this.studentName,
    this.attendanceId,
    this.status,
  });

  final String studentId;
  final String studentName;
  final String? attendanceId;
  final AttendanceStatus? status;

  RosterEntry withMark(String? attendanceId, AttendanceStatus? status) =>
      RosterEntry(
        studentId: studentId,
        studentName: studentName,
        attendanceId: attendanceId,
        status: status,
      );
}
