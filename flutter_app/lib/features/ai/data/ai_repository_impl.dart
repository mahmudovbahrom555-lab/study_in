import '../domain/entities/class_insights.dart';
import '../domain/repositories/ai_repository.dart';
import 'ai_api.dart';

class AiRepositoryImpl implements AiRepository {
  AiRepositoryImpl(this._api);

  final AiApi _api;

  @override
  Future<ClassInsights> classInsights(String groupId) async {
    final dto = await _api.classInsights(groupId);
    return dto.toDomain();
  }

  @override
  Future<StudentProgress> studentProgress(String groupId, String studentId) async {
    final dto = await _api.studentProgress(groupId, studentId);
    return dto.toDomain();
  }

  @override
  Future<void> submitQuestionFeedback(String questionId, {required bool accepted}) =>
      _api.submitQuestionFeedback(questionId, accepted: accepted);

  @override
  Future<({double rate, int total})> myAcceptanceRate() => _api.myAcceptanceRate();

  @override
  Future<TeacherGenerationStats> generationStats() async {
    final dto = await _api.generationStats();
    return dto.toDomain();
  }

  @override
  Future<void> recordRecommendationAction(
    String recId, {
    required String status,
    String action = '',
  }) =>
      _api.recordRecommendationAction(recId, status: status, action: action);

  @override
  Future<String> explainRecommendation(String recId) => _api.explainRecommendation(recId);

  @override
  Future<String> createDemoGroup() => _api.createDemoGroup();
}
