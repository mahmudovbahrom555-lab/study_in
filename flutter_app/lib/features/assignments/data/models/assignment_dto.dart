import '../../domain/entities/assignment.dart';
import '../../domain/entities/submission.dart';

Assignment assignmentFromJson(Map<String, dynamic> json) {
  final mine = json['my_submission'] as Map<String, dynamic>?;
  final stats = json['stats'] as Map<String, dynamic>?;
  return Assignment(
    id: json['id'] as String,
    groupId: json['group_id'] as String,
    teacherId: json['teacher_id'] as String,
    title: json['title'] as String,
    description: json['description'] as String?,
    dueDate: _date(json['due_date']),
    createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
    mySubmission: mine == null
        ? null
        : SubmissionSummary(
            grade: mine['grade'] as int?,
            submittedAt: DateTime.parse(mine['submitted_at'] as String).toLocal(),
          ),
    stats: stats == null
        ? null
        : SubmissionStats(
            submitted: stats['submitted'] as int,
            ungraded: stats['ungraded'] as int,
          ),
  );
}

Submission submissionFromJson(Map<String, dynamic> json) => Submission(
      id: json['id'] as String,
      studentId: json['student_id'] as String,
      studentName: json['student_name'] as String?,
      comment: json['comment'] as String?,
      grade: json['grade'] as int?,
      teacherNote: json['teacher_note'] as String?,
      files: ((json['attachments'] as List<dynamic>?) ?? const [])
          .map(
            (e) => attachedFileFromJson(
              e as Map<String, dynamic>,
              nameKey: 'filename',
            ),
          )
          .toList(),
      submittedAt: DateTime.parse(json['submitted_at'] as String).toLocal(),
      gradedAt: _date(json['graded_at']),
    );

/// Сдача отдаёт имя файла в `filename`, а POST /files — в `name`.
AttachedFile attachedFileFromJson(
  Map<String, dynamic> json, {
  String nameKey = 'name',
}) =>
    AttachedFile(
      id: json['id'] as String,
      name: json[nameKey] as String,
      mimeType: json['mime_type'] as String,
      sizeBytes: (json['size_bytes'] as num).toInt(),
      url: json['url'] as String,
    );

/// Сервер отдаёт UTC; на экране — время телефона.
DateTime? _date(Object? v) =>
    v == null ? null : DateTime.parse(v as String).toLocal();
