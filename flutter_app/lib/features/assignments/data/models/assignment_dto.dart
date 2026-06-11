import '../../domain/entities/assignment.dart';

class AssignmentDto {
  factory AssignmentDto.fromJson(Map<String, dynamic> json) {
    final sub = json['submission'] as Map<String, dynamic>?;
    return AssignmentDto(
      id: json['id'] as String,
      groupId: json['group_id'] as String,
      teacherId: json['teacher_id'] as String,
      title: json['title'] as String,
      description: json['description'] as String?,
      dueDate: json['due_date'] == null
          ? null
          : DateTime.parse(json['due_date'] as String),
      createdAt: DateTime.parse(json['created_at'] as String),
      submissionGrade: sub == null ? null : sub['grade'] as int?,
      isSubmitted: sub != null,
    );
  }

  const AssignmentDto({
    required this.id,
    required this.groupId,
    required this.teacherId,
    required this.title,
    this.description,
    this.dueDate,
    required this.createdAt,
    this.submissionGrade,
    this.isSubmitted = false,
  });

  final String id;
  final String groupId;
  final String teacherId;
  final String title;
  final String? description;
  final DateTime? dueDate;
  final DateTime createdAt;
  final int? submissionGrade;
  final bool isSubmitted;

  Assignment toDomain() => Assignment(
        id: id,
        groupId: groupId,
        teacherId: teacherId,
        title: title,
        description: description,
        dueDate: dueDate,
        createdAt: createdAt,
        submissionGrade: submissionGrade,
        isSubmitted: isSubmitted,
      );
}
