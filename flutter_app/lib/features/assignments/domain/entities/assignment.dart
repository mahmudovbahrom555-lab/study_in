class Assignment {
  const Assignment({
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
}
