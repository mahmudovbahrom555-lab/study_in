import '../../domain/entities/quiz.dart';

class OptionDto {
  factory OptionDto.fromJson(Map<String, dynamic> json) => OptionDto(
        id: json['id'] as String,
        questionId: json['question_id'] as String? ?? '',
        body: json['body'] as String,
        position: json['position'] as int? ?? 0,
        isCorrect: json['is_correct'] as bool?,
      );

  const OptionDto({
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

  Option toDomain() => Option(
        id: id,
        questionId: questionId,
        body: body,
        position: position,
        isCorrect: isCorrect,
      );
}

class QuestionDto {
  factory QuestionDto.fromJson(Map<String, dynamic> json) => QuestionDto(
        id: json['id'] as String,
        quizId: json['quiz_id'] as String? ?? '',
        body: json['body'] as String,
        explanation: json['explanation'] as String?,
        position: json['position'] as int? ?? 0,
        points: json['points'] as int? ?? 1,
        options: (json['options'] as List<dynamic>? ?? [])
            .map((e) => OptionDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  const QuestionDto({
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
  final List<OptionDto> options;

  Question toDomain() => Question(
        id: id,
        quizId: quizId,
        body: body,
        explanation: explanation,
        position: position,
        points: points,
        options: options.map((o) => o.toDomain()).toList(),
      );
}

class QuizDto {
  factory QuizDto.fromJson(Map<String, dynamic> json) => QuizDto(
        id: json['id'] as String,
        groupId: json['group_id'] as String,
        teacherId: json['teacher_id'] as String,
        title: json['title'] as String,
        description: json['description'] as String?,
        timeLimit: json['time_limit'] as int?,
        maxAttempts: json['max_attempts'] as int? ?? 1,
        isPublished: json['is_published'] as bool? ?? false,
        createdAt: DateTime.parse(json['created_at'] as String),
        openAt: json['open_at'] != null
            ? DateTime.parse(json['open_at'] as String)
            : null,
        closeAt: json['close_at'] != null
            ? DateTime.parse(json['close_at'] as String)
            : null,
        questions: (json['questions'] as List<dynamic>? ?? [])
            .map((e) => QuestionDto.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  const QuizDto({
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
  final List<QuestionDto> questions;

  Quiz toDomain() => Quiz(
        id: id,
        groupId: groupId,
        teacherId: teacherId,
        title: title,
        description: description,
        timeLimit: timeLimit,
        maxAttempts: maxAttempts,
        openAt: openAt,
        closeAt: closeAt,
        isPublished: isPublished,
        createdAt: createdAt,
        questions: questions.map((q) => q.toDomain()).toList(),
      );
}

class QuizAttemptDto {
  factory QuizAttemptDto.fromJson(Map<String, dynamic> json) => QuizAttemptDto(
        id: json['id'] as String,
        quizId: json['quiz_id'] as String,
        studentId: json['student_id'] as String,
        maxScore: json['max_score'] as int? ?? 0,
        startedAt: DateTime.parse(json['started_at'] as String),
        finishedAt: json['finished_at'] != null
            ? DateTime.parse(json['finished_at'] as String)
            : null,
        score: json['score'] as int?,
      );

  const QuizAttemptDto({
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

  QuizAttempt toDomain() => QuizAttempt(
        id: id,
        quizId: quizId,
        studentId: studentId,
        maxScore: maxScore,
        startedAt: startedAt,
        finishedAt: finishedAt,
        score: score,
      );
}
