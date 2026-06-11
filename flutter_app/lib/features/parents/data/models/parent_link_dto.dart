import '../../domain/entities/parent_link.dart';

class ParentLinkDto {
  factory ParentLinkDto.fromJson(Map<String, dynamic> json) => ParentLinkDto(
        id: json['id'] as String,
        parentId: json['parent_id'] as String,
        studentId: json['student_id'] as String,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  const ParentLinkDto({
    required this.id,
    required this.parentId,
    required this.studentId,
    required this.createdAt,
  });

  final String id;
  final String parentId;
  final String studentId;
  final DateTime createdAt;

  ParentLink toDomain() => ParentLink(
        id: id,
        parentId: parentId,
        studentId: studentId,
        createdAt: createdAt,
      );
}
