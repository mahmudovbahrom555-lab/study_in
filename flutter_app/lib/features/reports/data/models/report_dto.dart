import '../../domain/entities/report.dart';

class ParentReportDto {
  factory ParentReportDto.fromJson(Map<String, dynamic> j) => ParentReportDto(
        studentId: j['student_id'] as String,
        groupId: j['group_id'] as String,
        periodStart: j['period_start'] as String,
        periodEnd: j['period_end'] as String,
        attendancePct: (j['attendance_pct'] as num).toDouble(),
        quizScoreAvg: (j['quiz_score_avg'] as num).toDouble(),
        quizScoreDelta: (j['quiz_score_delta'] as num).toDouble(),
        masteryAvg: (j['mastery_avg'] as num).toDouble(),
        masteryDelta: (j['mastery_delta'] as num).toDouble(),
        homeworkCompletionPct: (j['homework_completion_pct'] as num).toDouble(),
        quizAttemptsCount: j['quiz_attempts_count'] as int,
        aiSummary: j['ai_summary'] as String?,
      );

  const ParentReportDto({
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

  ParentReport toDomain() => ParentReport(
        studentId: studentId,
        groupId: groupId,
        periodStart: periodStart,
        periodEnd: periodEnd,
        attendancePct: attendancePct,
        quizScoreAvg: quizScoreAvg,
        quizScoreDelta: quizScoreDelta,
        masteryAvg: masteryAvg,
        masteryDelta: masteryDelta,
        homeworkCompletionPct: homeworkCompletionPct,
        quizAttemptsCount: quizAttemptsCount,
        aiSummary: aiSummary,
      );
}

class RiskAlertDto {
  factory RiskAlertDto.fromJson(Map<String, dynamic> j) => RiskAlertDto(
        id: j['id'] as String,
        studentId: j['student_id'] as String,
        studentName: j['student_name'] as String,
        groupId: j['group_id'] as String,
        groupName: j['group_name'] as String,
        riskLevel: j['risk_level'] as String,
        triggerReason: j['trigger_reason'] as String,
        recommendation: j['recommendation'] as String,
        status: j['status'] as String,
        createdAt: DateTime.parse(j['created_at'] as String),
      );

  const RiskAlertDto({
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

  RiskAlert toDomain() => RiskAlert(
        id: id,
        studentId: studentId,
        studentName: studentName,
        groupId: groupId,
        groupName: groupName,
        riskLevel: riskLevel,
        triggerReason: triggerReason,
        recommendation: recommendation,
        status: status,
        createdAt: createdAt,
      );
}
