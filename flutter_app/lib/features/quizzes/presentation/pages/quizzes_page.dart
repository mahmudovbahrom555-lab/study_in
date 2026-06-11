import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../domain/entities/quiz.dart';
import '../providers/quizzes_provider.dart';
import '../widgets/create_quiz_sheet.dart';
import '../widgets/quiz_card.dart';

class QuizzesPage extends ConsumerStatefulWidget {
  const QuizzesPage({super.key, required this.groupId});

  final String groupId;

  @override
  ConsumerState<QuizzesPage> createState() => _QuizzesPageState();
}

class _QuizzesPageState extends ConsumerState<QuizzesPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(quizzesProvider(widget.groupId).notifier).load(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(quizzesProvider(widget.groupId));

    return Scaffold(
      appBar: AppBar(title: const Text('Тесты')),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCreate(context),
        child: const Icon(Icons.add),
      ),
      body: Builder(builder: (_) {
        if (state.isLoading && state.quizzes.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }
        if (state.error != null && state.quizzes.isEmpty) {
          return Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(state.error!, style: const TextStyle(color: Colors.red)),
                TextButton(
                  onPressed: () =>
                      ref.read(quizzesProvider(widget.groupId).notifier).load(),
                  child: const Text('Повторить'),
                ),
              ],
            ),
          );
        }
        if (state.quizzes.isEmpty) {
          return const Center(child: Text('Тестов пока нет'));
        }
        return RefreshIndicator(
          onRefresh: () =>
              ref.read(quizzesProvider(widget.groupId).notifier).load(),
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: state.quizzes.length,
            itemBuilder: (_, i) => QuizCard(
              quiz: state.quizzes[i],
              onTap: () => context.push(
                Routes.quiz(widget.groupId, state.quizzes[i].id),
              ),
              onTogglePublish: () => ref
                  .read(quizzesProvider(widget.groupId).notifier)
                  .togglePublish(state.quizzes[i]),
            ),
          ),
        );
      }),
    );
  }

  void _showCreate(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => CreateQuizSheet(
        onSubmit: ({
          required String title,
          String? description,
          int maxAttempts = 1,
          int? timeLimit,
        }) {
          ref.read(quizzesProvider(widget.groupId).notifier).createQuiz(
                title: title,
                description: description,
                maxAttempts: maxAttempts,
                timeLimit: timeLimit,
              );
        },
      ),
    );
  }
}
