/// Сдача ДЗ учеником (G10).
class Submission {
  const Submission({
    required this.id,
    required this.studentId,
    this.studentName,
    this.comment,
    this.grade,
    this.teacherNote,
    this.files = const [],
    required this.submittedAt,
    this.gradedAt,
  });

  final String id;
  final String studentId;
  final String? studentName;
  final String? comment;
  final int? grade;
  final String? teacherNote;
  final List<AttachedFile> files;
  final DateTime submittedAt;
  final DateTime? gradedAt;

  bool get isGraded => grade != null;
}

/// Файл на сервере: загруженный, но ещё не сданный, или уже в сдаче.
class AttachedFile {
  const AttachedFile({
    required this.id,
    required this.name,
    required this.mimeType,
    required this.sizeBytes,
    required this.url,
  });

  final String id;
  final String name;
  final String mimeType;
  final int sizeBytes;

  /// Подписанная ссылка, живёт час.
  final String url;

  bool get isImage => mimeType.startsWith('image/');
}
