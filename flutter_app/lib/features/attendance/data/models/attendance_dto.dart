import '../../domain/entities/attendance.dart';

class AttendanceDto {
  factory AttendanceDto.fromJson(Map<String, dynamic> json) => AttendanceDto(
        id: json['id'] as String,
        groupId: json['group_id'] as String,
        studentId: json['student_id'] as String,
        teacherId: json['teacher_id'] as String,
        lessonDate: json['lesson_date'] as String,
        status: AttendanceStatus.fromJson(json['status'] as String),
        note: json['note'] as String?,
        createdAt: DateTime.parse(json['created_at'] as String),
        updatedAt: DateTime.parse(json['updated_at'] as String),
      );

  const AttendanceDto({
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
  final String lessonDate;
  final AttendanceStatus status;
  final String? note;
  final DateTime createdAt;
  final DateTime updatedAt;

  Attendance toDomain() => Attendance(
        id: id,
        groupId: groupId,
        studentId: studentId,
        teacherId: teacherId,
        lessonDate: lessonDate,
        status: status,
        note: note,
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}
