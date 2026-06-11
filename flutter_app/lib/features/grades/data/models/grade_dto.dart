import '../../domain/entities/grade.dart';

class GradeDto {
  factory GradeDto.fromJson(Map<String, dynamic> json) => GradeDto(
        id: json['id'] as String,
        groupId: json['group_id'] as String,
        studentId: json['student_id'] as String,
        teacherId: json['teacher_id'] as String,
        subject: json['subject'] as String,
        value: (json['value'] as num).toDouble(),
        maxValue: (json['max_value'] as num?)?.toDouble() ?? 100.0,
        comment: json['comment'] as String?,
        gradedAt: DateTime.parse(json['graded_at'] as String),
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  const GradeDto({
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

  Grade toDomain() => Grade(
        id: id,
        groupId: groupId,
        studentId: studentId,
        teacherId: teacherId,
        subject: subject,
        value: value,
        maxValue: maxValue,
        comment: comment,
        gradedAt: gradedAt,
        createdAt: createdAt,
      );
}
