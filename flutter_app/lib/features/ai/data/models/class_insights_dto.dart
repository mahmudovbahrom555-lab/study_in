import '../../domain/entities/class_insights.dart';

class ClassInsightsDto {
  factory ClassInsightsDto.fromJson(Map<String, dynamic> j) => ClassInsightsDto(
        groupId: j['group_id'] as String,
        period: j['period'] as String? ?? 'last_30_days',
        studentCount: j['student_count'] as int? ?? 0,
        classWeakness: (j['class_weakness'] as List<dynamic>? ?? [])
            .map((e) => TopicWeaknessDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        students: (j['students'] as List<dynamic>? ?? [])
            .map((e) => StudentSummaryDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        quizStats: QuizStatsDto.fromJson(j['quiz_stats'] as Map<String, dynamic>? ?? {}),
        recommendations: (j['recommendations'] as List<dynamic>? ?? [])
            .map((e) => TeacherRecommendationDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  const ClassInsightsDto({
    required this.groupId,
    required this.period,
    required this.studentCount,
    required this.classWeakness,
    required this.students,
    required this.quizStats,
    required this.recommendations,
  });

  final String groupId;
  final String period;
  final int studentCount;
  final List<TopicWeaknessDto> classWeakness;
  final List<StudentSummaryDto> students;
  final QuizStatsDto quizStats;
  final List<TeacherRecommendationDto> recommendations;

  ClassInsights toDomain() => ClassInsights(
        groupId: groupId,
        period: period,
        studentCount: studentCount,
        classWeakness: classWeakness.map((e) => e.toDomain()).toList(),
        students: students.map((e) => e.toDomain()).toList(),
        quizStats: quizStats.toDomain(),
        recommendations: recommendations.map((e) => e.toDomain()).toList(),
      );
}

class TeacherRecommendationDto {
  factory TeacherRecommendationDto.fromJson(Map<String, dynamic> j) =>
      TeacherRecommendationDto(
        id: j['id'] as String? ?? '',
        priority: j['priority'] as int? ?? 3,
        action: j['action'] as String? ?? '',
        reason: j['reason'] as String? ?? '',
        topic: j['topic'] as String?,
        studentCount: j['student_count'] as int?,
      );

  const TeacherRecommendationDto({
    required this.id,
    required this.priority,
    required this.action,
    required this.reason,
    this.topic,
    this.studentCount,
  });

  final String id;
  final int priority;
  final String action;
  final String reason;
  final String? topic;
  final int? studentCount;

  TeacherRecommendation toDomain() => TeacherRecommendation(
        id: id,
        priority: priority,
        action: action,
        reason: reason,
        topic: topic,
        studentCount: studentCount,
      );
}

class TeacherGenerationStatsDto {
  factory TeacherGenerationStatsDto.fromJson(Map<String, dynamic> j) =>
      TeacherGenerationStatsDto(
        totalSessions: j['total_sessions'] as int? ?? 0,
        totalGenerated: j['total_generated'] as int? ?? 0,
        totalAccepted: j['total_accepted'] as int? ?? 0,
        totalEdited: j['total_edited'] as int? ?? 0,
        totalRejected: j['total_rejected'] as int? ?? 0,
        acceptanceRate: (j['acceptance_rate'] as num?)?.toDouble() ?? 0.0,
      );

  const TeacherGenerationStatsDto({
    required this.totalSessions,
    required this.totalGenerated,
    required this.totalAccepted,
    required this.totalEdited,
    required this.totalRejected,
    required this.acceptanceRate,
  });

  final int totalSessions;
  final int totalGenerated;
  final int totalAccepted;
  final int totalEdited;
  final int totalRejected;
  final double acceptanceRate;

  TeacherGenerationStats toDomain() => TeacherGenerationStats(
        totalSessions: totalSessions,
        totalGenerated: totalGenerated,
        totalAccepted: totalAccepted,
        totalEdited: totalEdited,
        totalRejected: totalRejected,
        acceptanceRate: acceptanceRate,
      );
}

class TopicWeaknessDto {
  factory TopicWeaknessDto.fromJson(Map<String, dynamic> j) => TopicWeaknessDto(
        topic: j['topic'] as String,
        avgAccuracy: (j['avg_accuracy'] as num?)?.toDouble() ?? 0.0,
        studentsStruggling: j['students_struggling'] as int? ?? 0,
        totalStudents: j['total_students'] as int? ?? 0,
        avgConfidence: (j['avg_confidence'] as num?)?.toDouble() ?? 0.5,
        avgConsistency: (j['avg_consistency'] as num?)?.toDouble() ?? 0.5,
      );

  const TopicWeaknessDto({
    required this.topic,
    required this.avgAccuracy,
    required this.studentsStruggling,
    required this.totalStudents,
    this.avgConfidence = 0.5,
    this.avgConsistency = 0.5,
  });

  final String topic;
  final double avgAccuracy;
  final int studentsStruggling;
  final int totalStudents;
  final double avgConfidence;
  final double avgConsistency;

  TopicWeakness toDomain() => TopicWeakness(
        topic: topic,
        avgAccuracy: avgAccuracy,
        studentsStruggling: studentsStruggling,
        totalStudents: totalStudents,
        avgConfidence: avgConfidence,
        avgConsistency: avgConsistency,
      );
}

class StudentSummaryDto {
  factory StudentSummaryDto.fromJson(Map<String, dynamic> j) => StudentSummaryDto(
        studentId: j['student_id'] as String,
        name: j['name'] as String,
        xpTotal: j['xp'] as int? ?? 0,
        streakDays: j['streak'] as int? ?? 0,
        topWeakness: j['top_weakness'] as String? ?? '',
        isAtRisk: j['is_at_risk'] as bool? ?? false,
        lastActive: j['last_active'] as String?,
      );

  const StudentSummaryDto({
    required this.studentId,
    required this.name,
    required this.xpTotal,
    required this.streakDays,
    required this.topWeakness,
    required this.isAtRisk,
    this.lastActive,
  });

  final String studentId;
  final String name;
  final int xpTotal;
  final int streakDays;
  final String topWeakness;
  final bool isAtRisk;
  final String? lastActive;

  StudentSummary toDomain() => StudentSummary(
        studentId: studentId,
        name: name,
        xpTotal: xpTotal,
        streakDays: streakDays,
        topWeakness: topWeakness,
        isAtRisk: isAtRisk,
        lastActive: lastActive,
      );
}

class QuizStatsDto {
  factory QuizStatsDto.fromJson(Map<String, dynamic> j) => QuizStatsDto(
        generated: j['generated'] as int? ?? 0,
        acceptanceRate: (j['acceptance_rate'] as num?)?.toDouble() ?? 0.0,
        totalAttempts: j['total_attempts'] as int? ?? 0,
        avgScore: (j['avg_score'] as num?)?.toDouble() ?? 0.0,
      );

  const QuizStatsDto({
    required this.generated,
    required this.acceptanceRate,
    required this.totalAttempts,
    required this.avgScore,
  });

  final int generated;
  final double acceptanceRate;
  final int totalAttempts;
  final double avgScore;

  QuizStats toDomain() => QuizStats(
        generated: generated,
        acceptanceRate: acceptanceRate,
        totalAttempts: totalAttempts,
        avgScore: avgScore,
      );
}

class StudentProgressDto {
  factory StudentProgressDto.fromJson(Map<String, dynamic> j) => StudentProgressDto(
        studentId: j['student_id'] as String,
        name: j['name'] as String,
        weakTopics: (j['weak_topics'] as List<dynamic>? ?? [])
            .map((e) => StudentTopicDetailDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        quizHistory: (j['quiz_history'] as List<dynamic>? ?? [])
            .map((e) => QuizAttemptSummaryDto.fromJson(e as Map<String, dynamic>))
            .toList(),
        skillLevels: (j['skill_levels'] as Map<String, dynamic>? ?? {}).map(
          (k, v) => MapEntry(k, SkillLevelDto.fromJson(v as Map<String, dynamic>)),
        ),
        gamification: j['gamification'] == null
            ? null
            : GamificationDto.fromJson(j['gamification'] as Map<String, dynamic>),
      );

  const StudentProgressDto({
    required this.studentId,
    required this.name,
    required this.weakTopics,
    required this.quizHistory,
    required this.skillLevels,
    this.gamification,
  });

  final String studentId;
  final String name;
  final List<StudentTopicDetailDto> weakTopics;
  final List<QuizAttemptSummaryDto> quizHistory;
  final Map<String, SkillLevelDto> skillLevels;
  final GamificationDto? gamification;

  StudentProgress toDomain() => StudentProgress(
        studentId: studentId,
        name: name,
        weakTopics: weakTopics.map((e) => e.toDomain()).toList(),
        quizHistory: quizHistory.map((e) => e.toDomain()).toList(),
        skillLevels: skillLevels.map((k, v) => MapEntry(k, v.toDomain())),
        gamification: gamification?.toDomain(),
      );
}

class StudentTopicDetailDto {
  factory StudentTopicDetailDto.fromJson(Map<String, dynamic> j) => StudentTopicDetailDto(
        topicId: j['topic_id'] as String,
        topicName: j['topic_name'] as String,
        accuracy: (j['accuracy'] as num?)?.toDouble() ?? 0.0,
        totalAnswers: j['total_answers'] as int? ?? 0,
        intervalDays: j['interval_days'] as int? ?? 1,
        confidenceScore: (j['confidence_score'] as num?)?.toDouble() ?? 0.5,
        consistencyScore: (j['consistency_score'] as num?)?.toDouble() ?? 0.5,
      );

  const StudentTopicDetailDto({
    required this.topicId,
    required this.topicName,
    required this.accuracy,
    required this.totalAnswers,
    required this.intervalDays,
    this.confidenceScore = 0.5,
    this.consistencyScore = 0.5,
  });

  final String topicId;
  final String topicName;
  final double accuracy;
  final int totalAnswers;
  final int intervalDays;
  final double confidenceScore;
  final double consistencyScore;

  StudentTopicDetail toDomain() => StudentTopicDetail(
        topicId: topicId,
        topicName: topicName,
        accuracy: accuracy,
        totalAnswers: totalAnswers,
        intervalDays: intervalDays,
        confidenceScore: confidenceScore,
        consistencyScore: consistencyScore,
      );
}

class QuizAttemptSummaryDto {
  factory QuizAttemptSummaryDto.fromJson(Map<String, dynamic> j) => QuizAttemptSummaryDto(
        quizTitle: j['quiz_title'] as String,
        score: (j['score'] as num?)?.toInt() ?? 0,
        maxScore: (j['max_score'] as num?)?.toInt() ?? 0,
        percentage: (j['percentage'] as num?)?.toDouble() ?? 0.0,
        finishedAt: DateTime.parse(j['finished_at'] as String),
      );

  const QuizAttemptSummaryDto({
    required this.quizTitle,
    required this.score,
    required this.maxScore,
    required this.percentage,
    required this.finishedAt,
  });

  final String quizTitle;
  final int score;
  final int maxScore;
  final double percentage;
  final DateTime finishedAt;

  QuizAttemptSummary toDomain() => QuizAttemptSummary(
        quizTitle: quizTitle,
        score: score,
        maxScore: maxScore,
        percentage: percentage,
        finishedAt: finishedAt,
      );
}

class SkillLevelDto {
  factory SkillLevelDto.fromJson(Map<String, dynamic> j) => SkillLevelDto(
        cefr: j['cefr'] as String? ?? '',
        score: (j['score'] as num?)?.toDouble() ?? 0.0,
        lastAssessed: DateTime.parse(j['last_assessed'] as String),
      );

  const SkillLevelDto({
    required this.cefr,
    required this.score,
    required this.lastAssessed,
  });

  final String cefr;
  final double score;
  final DateTime lastAssessed;

  SkillLevel toDomain() => SkillLevel(cefr: cefr, score: score, lastAssessed: lastAssessed);
}

class GamificationDto {
  factory GamificationDto.fromJson(Map<String, dynamic> j) => GamificationDto(
        xpTotal: j['xp_total'] as int? ?? 0,
        streakDays: j['streak_days'] as int? ?? 0,
        longestStreak: j['longest_streak'] as int? ?? 0,
        dailyGoalXp: j['daily_goal_xp'] as int? ?? 50,
      );

  const GamificationDto({
    required this.xpTotal,
    required this.streakDays,
    required this.longestStreak,
    required this.dailyGoalXp,
  });

  final int xpTotal;
  final int streakDays;
  final int longestStreak;
  final int dailyGoalXp;

  Gamification toDomain() => Gamification(
        xpTotal: xpTotal,
        streakDays: streakDays,
        longestStreak: longestStreak,
        dailyGoalXp: dailyGoalXp,
      );
}
