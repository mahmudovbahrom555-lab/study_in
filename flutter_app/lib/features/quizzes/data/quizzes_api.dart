import 'package:dio/dio.dart';

import 'models/quiz_dto.dart';

class QuizzesApi {
  QuizzesApi(this._dio);

  final Dio _dio;

  Future<List<QuizDto>> listGroupQuizzes(String groupId) async {
    final resp =
        await _dio.get<Map<String, dynamic>>('/groups/$groupId/quizzes');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => QuizDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<QuizDto> getQuiz(String groupId, String quizId) async {
    final resp = await _dio
        .get<Map<String, dynamic>>('/groups/$groupId/quizzes/$quizId');
    return QuizDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<QuizDto> createQuiz({
    required String groupId,
    required String title,
    String? description,
    int? timeLimit,
    int maxAttempts = 1,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes',
      data: {
        'title': title,
        if (description != null) 'description': description,
        if (timeLimit != null) 'time_limit': timeLimit,
        'max_attempts': maxAttempts,
      },
    );
    return QuizDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<QuizDto> publishQuiz(String groupId, String quizId) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes/$quizId/publish',
    );
    return QuizDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<QuizDto> unpublishQuiz(String groupId, String quizId) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes/$quizId/unpublish',
    );
    return QuizDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<QuestionDto> addQuestion({
    required String groupId,
    required String quizId,
    required String body,
    int points = 1,
    String? explanation,
    int position = 0,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes/$quizId/questions',
      data: {
        'body': body,
        'points': points,
        if (explanation != null) 'explanation': explanation,
        'position': position,
      },
    );
    return QuestionDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<OptionDto> addOption({
    required String groupId,
    required String quizId,
    required String questionId,
    required String body,
    required bool isCorrect,
    int position = 0,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes/$quizId/questions/$questionId/options',
      data: {
        'body': body,
        'is_correct': isCorrect,
        'position': position,
      },
    );
    return OptionDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<QuizAttemptDto> startAttempt(String groupId, String quizId) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes/$quizId/attempt',
    );
    return QuizAttemptDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<QuizResultDto> submitAttempt({
    required String groupId,
    required String quizId,
    required String attemptId,
    required Map<String, String> answers,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/quizzes/$quizId/attempt/$attemptId/submit',
      data: {'answers': answers},
    );
    return QuizResultDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }
}
