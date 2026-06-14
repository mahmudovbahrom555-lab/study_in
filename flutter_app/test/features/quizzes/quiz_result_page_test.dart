import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:repetapp/features/quizzes/domain/entities/quiz.dart';
import 'package:repetapp/features/quizzes/presentation/pages/quiz_result_page.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

QuizAttempt _attempt({required int score, required int maxScore}) {
  return QuizAttempt(
    id: 'atmp-1',
    quizId: 'quiz-1',
    studentId: 'stu-1',
    maxScore: maxScore,
    startedAt: DateTime(2024, 1, 1, 10),
    finishedAt: DateTime(2024, 1, 1, 10, 30),
    score: score,
  );
}

QuestionResult _qResult({
  required String questionBody,
  required String correctOptionId,
  required String correctBody,
  String? selectedOptionId,
  String? selectedBody,
  bool isCorrect = false,
}) {
  return QuestionResult(
    questionId: 'q-${questionBody.hashCode}',
    questionBody: questionBody,
    points: 1,
    correctOptionId: correctOptionId,
    correctBody: correctBody,
    isCorrect: isCorrect,
    selectedOptionId: selectedOptionId,
    selectedBody: selectedBody,
  );
}

Widget _wrap(Widget child) => MaterialApp(home: child);

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('QuizResultPage', () {
    final questionResults = [
      _qResult(
        questionBody: 'Question A',
        correctOptionId: 'opt-a',
        correctBody: 'Correct A',
        selectedOptionId: 'opt-a',
        selectedBody: 'Correct A',
        isCorrect: true,
      ),
      _qResult(
        questionBody: 'Question B',
        correctOptionId: 'opt-b',
        correctBody: 'Correct B',
        selectedOptionId: 'opt-x',
        selectedBody: 'Wrong B',
        isCorrect: false,
      ),
      _qResult(
        questionBody: 'Question C',
        correctOptionId: 'opt-c',
        correctBody: 'Correct C',
        // no selectedOptionId => unanswered
      ),
    ];

    final result80pct = QuizResult(
      attempt: _attempt(score: 8, maxScore: 10),
      questionResults: questionResults,
    );

    testWidgets('shows "80%" for score=8, maxScore=10', (tester) async {
      await tester.pumpWidget(_wrap(QuizResultPage(result: result80pct)));
      expect(find.text('80%'), findsOneWidget);
    });

    testWidgets('shows "+80 XP" for score=8 (xpEarned = score * 10)', (tester) async {
      await tester.pumpWidget(_wrap(QuizResultPage(result: result80pct)));
      expect(find.text('+80 XP'), findsOneWidget);
    });

    testWidgets('score header uses green color for >= 80%', (tester) async {
      await tester.pumpWidget(_wrap(QuizResultPage(result: result80pct)));

      // The percentage Text widget should exist and the _ScoreHeader card
      // background should use a green-ish color.  We confirm the widget is
      // present; color is verified via the Card's decoration in the widget tree.
      final pctFinder = find.text('80%');
      expect(pctFinder, findsOneWidget);

      // The text colour is set to Colors.green in the widget for pct >= 0.8.
      final textWidget = tester.widget<Text>(pctFinder);
      expect(
        textWidget.style?.color,
        Colors.green,
        reason: 'percentage text must be green when score >= 80%',
      );
    });

    testWidgets('renders a card for each question result', (tester) async {
      await tester.pumpWidget(_wrap(QuizResultPage(result: result80pct)));

      // Each _QuestionResultCard contains the question body as a Text
      expect(find.textContaining('Question A'), findsOneWidget);
      expect(find.textContaining('Question B'), findsOneWidget);
      expect(find.textContaining('Question C'), findsOneWidget);
    });

    testWidgets('orange color for score between 60-79%', (tester) async {
      final result70 = QuizResult(
        attempt: _attempt(score: 7, maxScore: 10),
        questionResults: questionResults,
      );
      await tester.pumpWidget(_wrap(QuizResultPage(result: result70)));

      final pctFinder = find.text('70%');
      expect(pctFinder, findsOneWidget);
      final textWidget = tester.widget<Text>(pctFinder);
      expect(
        textWidget.style?.color,
        Colors.orange,
        reason: 'percentage text must be orange for 60-79%',
      );
    });

    testWidgets('red color for score below 60%', (tester) async {
      final result50 = QuizResult(
        attempt: _attempt(score: 5, maxScore: 10),
        questionResults: questionResults,
      );
      await tester.pumpWidget(_wrap(QuizResultPage(result: result50)));

      final pctFinder = find.text('50%');
      expect(pctFinder, findsOneWidget);
      final textWidget = tester.widget<Text>(pctFinder);
      expect(
        textWidget.style?.color,
        Colors.red,
        reason: 'percentage text must be red for < 60%',
      );
    });

    testWidgets('zero maxScore does not throw — shows 0%', (tester) async {
      final resultZero = QuizResult(
        attempt: _attempt(score: 0, maxScore: 0),
        questionResults: const [],
      );
      await tester.pumpWidget(_wrap(QuizResultPage(result: resultZero)));
      expect(find.text('0%'), findsOneWidget);
      expect(find.text('+0 XP'), findsOneWidget);
    });
  });
}
