import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/quizzes_api.dart';
import '../../data/quizzes_repository_impl.dart';
import '../../domain/entities/quiz.dart';
import '../../domain/repositories/quizzes_repository.dart';

final quizzesRepositoryProvider =
    Provider.family<QuizzesRepository, String>((ref, groupId) {
  final dio = ref.watch(dioProvider);
  final impl = QuizzesRepositoryImpl(QuizzesApi(dio));
  impl.setGroupId(groupId);
  return impl;
});

// ─── Quiz list ────────────────────────────────────────────────────────────────

class QuizzesState {
  const QuizzesState({
    this.quizzes = const [],
    this.isLoading = false,
    this.error,
  });

  final List<Quiz> quizzes;
  final bool isLoading;
  final String? error;

  QuizzesState copyWith({
    List<Quiz>? quizzes,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      QuizzesState(
        quizzes: quizzes ?? this.quizzes,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

class QuizzesNotifier extends StateNotifier<QuizzesState> {
  QuizzesNotifier(this._repo, this._groupId) : super(const QuizzesState());

  final QuizzesRepository _repo;
  final String _groupId;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final qs = await _repo.listGroupQuizzes(_groupId);
      state = state.copyWith(isLoading: false, quizzes: qs);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> createQuiz({
    required String title,
    String? description,
    int maxAttempts = 1,
    int? timeLimit,
  }) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final q = await _repo.createQuiz(
        groupId: _groupId,
        title: title,
        description: description,
        maxAttempts: maxAttempts,
        timeLimit: timeLimit,
      );
      state = state.copyWith(
        isLoading: false,
        quizzes: [q, ...state.quizzes],
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> togglePublish(Quiz quiz) async {
    try {
      final updated = quiz.isPublished
          ? await _repo.unpublishQuiz(quiz.id)
          : await _repo.publishQuiz(quiz.id);
      state = state.copyWith(
        quizzes: state.quizzes
            .map((q) => q.id == quiz.id ? updated : q)
            .toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

final quizzesProvider = StateNotifierProvider.family<QuizzesNotifier,
    QuizzesState, String>((ref, groupId) {
  return QuizzesNotifier(ref.watch(quizzesRepositoryProvider(groupId)), groupId);
});

// ─── Quiz detail ──────────────────────────────────────────────────────────────

class QuizDetailNotifier extends StateNotifier<AsyncValue<Quiz>> {
  QuizDetailNotifier(this._repo, this._groupId, this._quizId)
      : super(const AsyncValue.loading()) {
    load();
  }

  final QuizzesRepository _repo;
  final String _groupId;
  final String _quizId;

  Future<void> load() async {
    state = const AsyncValue.loading();
    try {
      final q = await _repo.getQuiz(_quizId);
      state = AsyncValue.data(q);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}

final quizDetailProvider = StateNotifierProvider.family<QuizDetailNotifier,
    AsyncValue<Quiz>, ({String groupId, String quizId})>(
  (ref, params) => QuizDetailNotifier(
    ref.watch(quizzesRepositoryProvider(params.groupId)),
    params.groupId,
    params.quizId,
  ),
);

// ─── Attempt ──────────────────────────────────────────────────────────────────

class AttemptState {
  const AttemptState({
    this.attempt,
    this.answers = const {},
    this.isSubmitting = false,
    this.error,
  });

  final QuizAttempt? attempt;
  final Map<String, String> answers; // questionId → optionId
  final bool isSubmitting;
  final String? error;

  AttemptState copyWith({
    QuizAttempt? attempt,
    Map<String, String>? answers,
    bool? isSubmitting,
    String? error,
    bool clearError = false,
  }) =>
      AttemptState(
        attempt: attempt ?? this.attempt,
        answers: answers ?? this.answers,
        isSubmitting: isSubmitting ?? this.isSubmitting,
        error: clearError ? null : (error ?? this.error),
      );
}

class AttemptNotifier extends StateNotifier<AttemptState> {
  AttemptNotifier(this._repo) : super(const AttemptState());

  final QuizzesRepository _repo;

  Future<void> start(String quizId) async {
    try {
      final attempt = await _repo.startAttempt(quizId);
      state = state.copyWith(attempt: attempt, answers: {});
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  void selectAnswer(String questionId, String optionId) {
    state = state.copyWith(
      answers: {...state.answers, questionId: optionId},
    );
  }

  Future<QuizAttempt?> submit(String quizId) async {
    final attempt = state.attempt;
    if (attempt == null) return null;
    state = state.copyWith(isSubmitting: true, clearError: true);
    try {
      final result = await _repo.submitAttempt(
        quizId: quizId,
        attemptId: attempt.id,
        answers: state.answers,
      );
      state = state.copyWith(isSubmitting: false, attempt: result);
      return result;
    } catch (e) {
      state = state.copyWith(isSubmitting: false, error: e.toString());
      return null;
    }
  }
}

final attemptProvider =
    StateNotifierProvider.autoDispose<AttemptNotifier, AttemptState>((ref) {
  // groupId must be set on the repo before use; caller uses quizzesRepositoryProvider(groupId)
  // We use a simple provider here — caller passes groupId-scoped repo via parameter
  throw UnimplementedError('use attemptProviderFamily');
});

final attemptProviderFamily = StateNotifierProvider.autoDispose
    .family<AttemptNotifier, AttemptState, String>((ref, groupId) {
  return AttemptNotifier(ref.watch(quizzesRepositoryProvider(groupId)));
});
