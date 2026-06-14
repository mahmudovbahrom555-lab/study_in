class ParentReport {
  const ParentReport({
    required this.studentId,
    required this.groupId,
    required this.periodStart,
    required this.periodEnd,
    required this.attendancePct,
    required this.quizScoreAvg,
    required this.quizScoreDelta,
    required this.masteryAvg,
    required this.masteryDelta,
    required this.homeworkCompletionPct,
    required this.quizAttemptsCount,
    this.aiSummary,
  });

  final String studentId;
  final String groupId;
  final String periodStart;
  final String periodEnd;
  final double attendancePct;
  final double quizScoreAvg;
  final double quizScoreDelta;
  final double masteryAvg;
  final double masteryDelta;
  final double homeworkCompletionPct;
  final int quizAttemptsCount;
  final String? aiSummary;
}

class RiskAlert {
  const RiskAlert({
    required this.id,
    required this.studentId,
    required this.studentName,
    required this.groupId,
    required this.groupName,
    required this.riskLevel,
    required this.triggerReason,
    required this.recommendation,
    required this.status,
    required this.createdAt,
  });

  final String id;
  final String studentId;
  final String studentName;
  final String groupId;
  final String groupName;
  final String riskLevel;
  final String triggerReason;
  final String recommendation;
  final String status;
  final DateTime createdAt;

  bool get isHigh => riskLevel == 'high';
  bool get isMedium => riskLevel == 'medium';
  bool get isResolved => status == 'resolved';
}
