import '../entities/class_insights.dart';

abstract class AiRepository {
  Future<ClassInsights> classInsights(String groupId);
  Future<StudentProgress> studentProgress(String groupId, String studentId);
  Future<void> submitQuestionFeedback(String questionId, {required bool accepted});
  Future<({double rate, int total})> myAcceptanceRate();
  Future<TeacherGenerationStats> generationStats();
  Future<void> recordRecommendationAction(String recId, {required String status, String action});
  Future<String> explainRecommendation(String recId);
  Future<String> createDemoGroup();
  Future<Gamification> myGamification();
}
