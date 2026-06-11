import 'package:flutter/foundation.dart';

@immutable
class Quiz {
  const Quiz({
    required this.id,
    required this.groupId,
    required this.teacherId,
    required this.title,
    required this.isPublished,
    required this.maxAttempts,
    required this.createdAt,
    this.description,
    this.timeLimit,
    this.openAt,
    this.closeAt,
    this.questions = const [],
  });

  final String id;
  final String groupId;
  final String teacherId;
  final String title;
  final String? description;
  final int? timeLimit;
  final int maxAttempts;
  final DateTime? openAt;
  final DateTime? closeAt;
  final bool isPublished;
  final DateTime createdAt;
  final List<Question> questions;
}

@immutable
class Question {
  const Question({
    required this.id,
    required this.quizId,
    required this.body,
    required this.position,
    required this.points,
    this.explanation,
    this.options = const [],
  });

  final String id;
  final String quizId;
  final String body;
  final String? explanation;
  final int position;
  final int points;
  final List<Option> options;
}

@immutable
class Option {
  const Option({
    required this.id,
    required this.questionId,
    required this.body,
    required this.position,
    this.isCorrect,
  });

  final String id;
  final String questionId;
  final String body;
  final int position;
  final bool? isCorrect;
}

@immutable
class QuizAttempt {
  const QuizAttempt({
    required this.id,
    required this.quizId,
    required this.studentId,
    required this.maxScore,
    required this.startedAt,
    this.finishedAt,
    this.score,
  });

  final String id;
  final String quizId;
  final String studentId;
  final int maxScore;
  final DateTime startedAt;
  final DateTime? finishedAt;
  final int? score;

  bool get isFinished => finishedAt != null;
}
