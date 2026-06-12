import 'package:dio/dio.dart';

import 'models/class_insights_dto.dart';

class AiApi {
  AiApi(this._dio);

  final Dio _dio;

  Future<ClassInsightsDto> classInsights(String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>('/groups/$groupId/ai-insights');
    return ClassInsightsDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<StudentProgressDto> studentProgress(String groupId, String studentId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/students/$studentId/progress',
    );
    return StudentProgressDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> submitQuestionFeedback(String questionId, {required bool accepted}) =>
      _dio.post<void>('/questions/$questionId/feedback', data: {'accepted': accepted});

  Future<({double rate, int total})> myAcceptanceRate() async {
    final resp = await _dio.get<Map<String, dynamic>>('/me/acceptance-rate');
    final d = resp.data!['data'] as Map<String, dynamic>;
    return (
      rate: (d['acceptance_rate'] as num).toDouble(),
      total: d['total_questions'] as int,
    );
  }
}
