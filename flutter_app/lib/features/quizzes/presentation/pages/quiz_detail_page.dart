import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../domain/entities/quiz.dart';
import '../providers/quizzes_provider.dart';

class QuizDetailPage extends ConsumerWidget {
  const QuizDetailPage({
    super.key,
    required this.groupId,
    required this.quizId,
  });

  final String groupId;
  final String quizId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(
      quizDetailProvider((groupId: groupId, quizId: quizId)),
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Тест')),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (quiz) => _QuizBody(quiz: quiz, groupId: groupId),
      ),
    );
  }
}

class _QuizBody extends StatelessWidget {
  const _QuizBody({required this.quiz, required this.groupId});

  final Quiz quiz;
  final String groupId;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(quiz.title, style: theme.textTheme.headlineSmall),
        if (quiz.description != null) ...[
          const SizedBox(height: 8),
          Text(quiz.description!, style: theme.textTheme.bodyMedium),
        ],
        const SizedBox(height: 12),
        Wrap(
          spacing: 8,
          children: [
            Chip(
              avatar: const Icon(Icons.repeat, size: 16),
              label: Text('${quiz.maxAttempts} попыток'),
            ),
            if (quiz.timeLimit != null)
              Chip(
                avatar: const Icon(Icons.timer_outlined, size: 16),
                label: Text('${quiz.timeLimit} мин'),
              ),
            Chip(
              avatar: Icon(
                quiz.isPublished ? Icons.visibility : Icons.visibility_off,
                size: 16,
                color: quiz.isPublished ? Colors.green : Colors.orange,
              ),
              label: Text(quiz.isPublished ? 'Опубликован' : 'Черновик'),
            ),
          ],
        ),
        const Divider(height: 32),
        if (quiz.questions.isEmpty)
          const Text('Вопросов нет')
        else ...[
          Text(
            'Вопросы (${quiz.questions.length})',
            style: theme.textTheme.titleMedium,
          ),
          const SizedBox(height: 12),
          ...quiz.questions.asMap().entries.map(
                (e) => _QuestionTile(
                  index: e.key + 1,
                  question: e.value,
                ),
              ),
        ],
        const SizedBox(height: 24),
        if (quiz.isPublished)
          FilledButton.icon(
            icon: const Icon(Icons.play_arrow),
            label: const Text('Начать тест'),
            onPressed: () => context.push(
              Routes.quizAttempt(groupId, quiz.id),
            ),
          ),
      ],
    );
  }
}

class _QuestionTile extends StatelessWidget {
  const _QuestionTile({required this.index, required this.question});

  final int index;
  final Question question;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  radius: 12,
                  child: Text('$index', style: const TextStyle(fontSize: 12)),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    question.body,
                    style: theme.textTheme.bodyLarge,
                  ),
                ),
                Text(
                  '${question.points} б.',
                  style: theme.textTheme.labelSmall,
                ),
              ],
            ),
            if (question.options.isNotEmpty) ...[
              const SizedBox(height: 8),
              ...question.options.map(
                (o) => Padding(
                  padding: const EdgeInsets.only(left: 32, top: 4),
                  child: Row(
                    children: [
                      Icon(
                        o.isCorrect == true
                            ? Icons.check_circle
                            : Icons.radio_button_unchecked,
                        size: 16,
                        color: o.isCorrect == true
                            ? Colors.green
                            : Colors.grey,
                      ),
                      const SizedBox(width: 6),
                      Expanded(child: Text(o.body)),
                    ],
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
