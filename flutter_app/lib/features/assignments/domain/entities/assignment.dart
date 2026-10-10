class Assignment {
  const Assignment({
    required this.id,
    required this.groupId,
    required this.teacherId,
    required this.title,
    this.description,
    this.dueDate,
    required this.createdAt,
    this.mySubmission,
    this.stats,
  });

  final String id;
  final String groupId;
  final String teacherId;
  final String title;
  final String? description;
  final DateTime? dueDate;
  final DateTime createdAt;

  /// Ученику — своя сдача; null — ещё не сдавал.
  final SubmissionSummary? mySubmission;

  /// Репетитору — сколько работ сдано и сколько ждут проверки.
  final SubmissionStats? stats;

  bool get isSubmitted => mySubmission != null;

  bool get isOverdue =>
      dueDate != null && !isSubmitted && dueDate!.isBefore(DateTime.now());
}

class SubmissionSummary {
  const SubmissionSummary({this.grade, required this.submittedAt});

  final int? grade;
  final DateTime submittedAt;
}

class SubmissionStats {
  const SubmissionStats({required this.submitted, required this.ungraded});

  final int submitted;
  final int ungraded;
}
