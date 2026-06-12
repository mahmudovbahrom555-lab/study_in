import '../domain/entities/quiz.dart';
import '../domain/repositories/quizzes_repository.dart';
import 'quizzes_api.dart';

class QuizzesRepositoryImpl implements QuizzesRepository {
  QuizzesRepositoryImpl(this._api);

  final QuizzesApi _api;

  // groupId is needed by the API but the domain interface doesn't carry it,
  // so we parse it from the quiz's groupId after fetching.
  // For list/create we require it explicitly via a workaround:
  // store the groupId that was used for the last list call.
  String _currentGroupId = '';

  void setGroupId(String groupId) => _currentGroupId = groupId;

  @override
  Future<List<Quiz>> listGroupQuizzes(String groupId) async {
    _currentGroupId = groupId;
    final dtos = await _api.listGroupQuizzes(groupId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<Quiz> getQuiz(String quizId) async {
    final dto = await _api.getQuiz(_currentGroupId, quizId);
    return dto.toDomain();
  }

  @override
  Future<Quiz> createQuiz({
    required String groupId,
    required String title,
    String? description,
    int? timeLimit,
    int maxAttempts = 1,
  }) async {
    _currentGroupId = groupId;
    final dto = await _api.createQuiz(
      groupId: groupId,
      title: title,
      description: description,
      timeLimit: timeLimit,
      maxAttempts: maxAttempts,
    );
    return dto.toDomain();
  }

  @override
  Future<Quiz> publishQuiz(String quizId) async {
    final dto = await _api.publishQuiz(_currentGroupId, quizId);
    return dto.toDomain();
  }

  @override
  Future<Quiz> unpublishQuiz(String quizId) async {
    final dto = await _api.unpublishQuiz(_currentGroupId, quizId);
    return dto.toDomain();
  }

  @override
  Future<Question> addQuestion({
    required String quizId,
    required String body,
    int points = 1,
    String? explanation,
    int position = 0,
  }) async {
    final dto = await _api.addQuestion(
      groupId: _currentGroupId,
      quizId: quizId,
      body: body,
      points: points,
      explanation: explanation,
      position: position,
    );
    return dto.toDomain();
  }

  @override
  Future<Option> addOption({
    required String questionId,
    required String body,
    required bool isCorrect,
    int position = 0,
  }) async {
    // questionId is enough to locate the question; quizId comes from context
    final dto = await _api.addOption(
      groupId: _currentGroupId,
      quizId: _currentQuizId,
      questionId: questionId,
      body: body,
      isCorrect: isCorrect,
      position: position,
    );
    return dto.toDomain();
  }

  String _currentQuizId = '';

  void setQuizId(String quizId) => _currentQuizId = quizId;

  @override
  Future<QuizAttempt> startAttempt(String quizId) async {
    _currentQuizId = quizId;
    final dto = await _api.startAttempt(_currentGroupId, quizId);
    return dto.toDomain();
  }

  @override
  Future<QuizResult> submitAttempt({
    required String quizId,
    required String attemptId,
    required Map<String, String> answers,
  }) async {
    final dto = await _api.submitAttempt(
      groupId: _currentGroupId,
      quizId: quizId,
      attemptId: attemptId,
      answers: answers,
    );
    return dto.toDomain();
  }
}
