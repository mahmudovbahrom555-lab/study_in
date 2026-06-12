import '../entities/quiz.dart';

abstract class QuizzesRepository {
  Future<List<Quiz>> listGroupQuizzes(String groupId);
  Future<Quiz> getQuiz(String quizId);
  Future<Quiz> createQuiz({
    required String groupId,
    required String title,
    String? description,
    int? timeLimit,
    int maxAttempts,
  });
  Future<Quiz> publishQuiz(String quizId);
  Future<Quiz> unpublishQuiz(String quizId);
  Future<Question> addQuestion({
    required String quizId,
    required String body,
    int points,
    String? explanation,
    int position,
  });
  Future<Option> addOption({
    required String questionId,
    required String body,
    required bool isCorrect,
    int position,
  });
  Future<QuizAttempt> startAttempt(String quizId);
  Future<QuizResult> submitAttempt({
    required String quizId,
    required String attemptId,
    required Map<String, String> answers,
  });
}
