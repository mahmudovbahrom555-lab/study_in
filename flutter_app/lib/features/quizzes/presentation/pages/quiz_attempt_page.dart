import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/quiz.dart';
import '../providers/quizzes_provider.dart';
import 'quiz_result_page.dart';

class QuizAttemptPage extends ConsumerStatefulWidget {
  const QuizAttemptPage({
    super.key,
    required this.groupId,
    required this.quizId,
  });

  final String groupId;
  final String quizId;

  @override
  ConsumerState<QuizAttemptPage> createState() => _QuizAttemptPageState();
}

class _QuizAttemptPageState extends ConsumerState<QuizAttemptPage> {
  Quiz? _quiz;
  bool _loadingQuiz = true;
  bool _started = false;

  @override
  void initState() {
    super.initState();
    _loadQuiz();
  }

  Future<void> _loadQuiz() async {
    final repo = ref.read(quizzesRepositoryProvider(widget.groupId));
    try {
      final q = await repo.getQuiz(widget.quizId);
      if (mounted) setState(() { _quiz = q; _loadingQuiz = false; });
    } catch (_) {
      if (mounted) setState(() => _loadingQuiz = false);
    }
  }

  Future<void> _startAttempt() async {
    await ref
        .read(attemptProviderFamily(widget.groupId).notifier)
        .start(widget.quizId);
    if (mounted) setState(() => _started = true);
  }

  Future<void> _submit() async {
    final result = await ref
        .read(attemptProviderFamily(widget.groupId).notifier)
        .submit(widget.quizId);
    if (result != null && mounted) {
      await Navigator.of(context).push(
        MaterialPageRoute<void>(
          builder: (_) => QuizResultPage(result: result),
        ),
      );
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loadingQuiz) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    if (_quiz == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Тест')),
        body: const Center(child: Text('Тест не найден')),
      );
    }

    if (!_started) {
      return _StartScreen(quiz: _quiz!, onStart: _startAttempt);
    }

    return _AttemptScreen(
      quiz: _quiz!,
      groupId: widget.groupId,
      onSubmit: _submit,
    );
  }
}

class _StartScreen extends StatelessWidget {
  const _StartScreen({required this.quiz, required this.onStart});

  final Quiz quiz;
  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(quiz.title)),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              quiz.title,
              style: Theme.of(context).textTheme.headlineMedium,
              textAlign: TextAlign.center,
            ),
            if (quiz.description != null) ...[
              const SizedBox(height: 12),
              Text(quiz.description!, textAlign: TextAlign.center),
            ],
            const SizedBox(height: 24),
            if (quiz.timeLimit != null)
              _InfoRow(
                icon: Icons.timer_outlined,
                label: 'Ограничение времени: ${quiz.timeLimit} мин',
              ),
            _InfoRow(
              icon: Icons.quiz_outlined,
              label: 'Вопросов: ${quiz.questions.length}',
            ),
            _InfoRow(
              icon: Icons.repeat,
              label: 'Попыток: ${quiz.maxAttempts}',
            ),
            const SizedBox(height: 40),
            FilledButton.icon(
              icon: const Icon(Icons.play_arrow),
              label: const Text('Начать'),
              onPressed: onStart,
            ),
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Icon(icon, size: 20, color: Theme.of(context).colorScheme.primary),
          const SizedBox(width: 12),
          Text(label),
        ],
      ),
    );
  }
}

class _AttemptScreen extends ConsumerWidget {
  const _AttemptScreen({
    required this.quiz,
    required this.groupId,
    required this.onSubmit,
  });

  final Quiz quiz;
  final String groupId;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(attemptProviderFamily(groupId));
    final answered = state.answers.length;
    final total = quiz.questions.length;

    return Scaffold(
      appBar: AppBar(
        title: Text(quiz.title),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(4),
          child: LinearProgressIndicator(
            value: total == 0 ? 0 : answered / total,
          ),
        ),
      ),
      body: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: quiz.questions.length,
        itemBuilder: (_, i) {
          final q = quiz.questions[i];
          final selected = state.answers[q.id];
          return _QuestionCard(
            index: i + 1,
            question: q,
            selectedOptionId: selected,
            onSelect: (optionId) => ref
                .read(attemptProviderFamily(groupId).notifier)
                .selectAnswer(q.id, optionId),
          );
        },
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: FilledButton(
            onPressed: state.isSubmitting ? null : onSubmit,
            child: state.isSubmitting
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Text('Сдать ($answered / $total)'),
          ),
        ),
      ),
    );
  }
}

class _QuestionCard extends StatelessWidget {
  const _QuestionCard({
    required this.index,
    required this.question,
    required this.selectedOptionId,
    required this.onSelect,
  });

  final int index;
  final Question question;
  final String? selectedOptionId;
  final void Function(String optionId) onSelect;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '$index. ${question.body}',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            ...question.options.map(
              (o) => RadioListTile<String>(
                value: o.id,
                groupValue: selectedOptionId,
                onChanged: (_) => onSelect(o.id),
                title: Text(o.body),
                contentPadding: EdgeInsets.zero,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
