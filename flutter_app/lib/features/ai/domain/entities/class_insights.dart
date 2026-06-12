class ClassInsights {
  const ClassInsights({
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
  final List<TopicWeakness> classWeakness;
  final List<StudentSummary> students;
  final QuizStats quizStats;
  final List<TeacherRecommendation> recommendations;
}

class TopicWeakness {
  const TopicWeakness({
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
}

class StudentSummary {
  const StudentSummary({
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
}

class QuizStats {
  const QuizStats({
    required this.generated,
    required this.acceptanceRate,
    required this.totalAttempts,
    required this.avgScore,
  });

  final int generated;
  final double acceptanceRate;
  final int totalAttempts;
  final double avgScore;
}

class TeacherRecommendation {
  const TeacherRecommendation({
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
}

class TeacherGenerationStats {
  const TeacherGenerationStats({
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
}

class StudentProgress {
  const StudentProgress({
    required this.studentId,
    required this.name,
    required this.weakTopics,
    required this.quizHistory,
    required this.skillLevels,
    this.gamification,
  });

  final String studentId;
  final String name;
  final List<StudentTopicDetail> weakTopics;
  final List<QuizAttemptSummary> quizHistory;
  final Map<String, SkillLevel> skillLevels;
  final Gamification? gamification;
}

class StudentTopicDetail {
  const StudentTopicDetail({
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
}

class QuizAttemptSummary {
  const QuizAttemptSummary({
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
}

class SkillLevel {
  const SkillLevel({required this.cefr, required this.score, required this.lastAssessed});

  final String cefr;
  final double score;
  final DateTime lastAssessed;
}

class Gamification {
  const Gamification({
    required this.xpTotal,
    required this.streakDays,
    required this.longestStreak,
    required this.dailyGoalXp,
  });

  final int xpTotal;
  final int streakDays;
  final int longestStreak;
  final int dailyGoalXp;
}
