import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../domain/entities/class_insights.dart';
import '../providers/ai_insights_provider.dart';

class StudentProgressPage extends ConsumerWidget {
  const StudentProgressPage({
    super.key,
    required this.groupId,
    required this.studentId,
  });

  final String groupId;
  final String studentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(
      studentProgressProvider((groupId: groupId, studentId: studentId)),
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Прогресс студента')),
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (progress) => _ProgressBody(progress: progress),
      ),
    );
  }
}

class _ProgressBody extends StatelessWidget {
  const _ProgressBody({required this.progress});

  final StudentProgress progress;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _GamificationCard(progress: progress),
        const SizedBox(height: 16),
        if (progress.skillLevels.isNotEmpty) ...[
          Text('Навыки (CEFR)', style: theme.textTheme.titleMedium),
          const SizedBox(height: 8),
          _SkillGrid(skills: progress.skillLevels),
          const SizedBox(height: 16),
        ],
        Text('Слабые темы', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        if (progress.weakTopics.isEmpty)
          const Text('Нет данных')
        else
          ...progress.weakTopics.map((t) => _TopicRow(topic: t)),
        const SizedBox(height: 16),
        Text('История тестов', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        if (progress.quizHistory.isEmpty)
          const Text('Тесты ещё не проходились')
        else
          ...progress.quizHistory.map((a) => _AttemptTile(attempt: a)),
      ],
    );
  }
}

class _GamificationCard extends StatelessWidget {
  const _GamificationCard({required this.progress});

  final StudentProgress progress;

  @override
  Widget build(BuildContext context) {
    final g = progress.gamification;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(progress.name, style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 8),
            Row(
              children: [
                _Chip(
                  label: '${g?.xpTotal ?? 0} XP',
                  icon: Icons.star,
                  color: Colors.amber,
                ),
                const SizedBox(width: 8),
                _Chip(
                  label: '🔥 ${g?.streakDays ?? 0} дн.',
                  icon: Icons.local_fire_department,
                  color: Colors.orange,
                ),
                const SizedBox(width: 8),
                _Chip(
                  label: 'Рекорд: ${g?.longestStreak ?? 0}',
                  icon: Icons.emoji_events,
                  color: Colors.green,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({required this.label, required this.icon, required this.color});

  final String label;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Chip(
      avatar: Icon(icon, color: color, size: 16),
      label: Text(label, style: const TextStyle(fontSize: 12)),
      visualDensity: VisualDensity.compact,
    );
  }
}

class _SkillGrid extends StatelessWidget {
  const _SkillGrid({required this.skills});

  final Map<String, SkillLevel> skills;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: skills.entries.map((e) {
        return Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.primaryContainer,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(e.key, style: const TextStyle(fontWeight: FontWeight.bold)),
              Text(
                e.value.cefr.isNotEmpty ? e.value.cefr : '—',
                style: const TextStyle(fontSize: 18),
              ),
              Text(
                '${(e.value.score).toStringAsFixed(0)}%',
                style: const TextStyle(fontSize: 12),
              ),
            ],
          ),
        );
      }).toList(),
    );
  }
}

class _TopicRow extends StatelessWidget {
  const _TopicRow({required this.topic});

  final StudentTopicDetail topic;

  @override
  Widget build(BuildContext context) {
    final pct = (topic.accuracy * 100).toStringAsFixed(0);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          Expanded(child: Text(topic.topicName)),
          Text(
            '$pct%',
            style: TextStyle(
              color: topic.accuracy < 0.6 ? Colors.red : Colors.green,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(width: 8),
          Text(
            '(${topic.totalAnswers} отв.)',
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ],
      ),
    );
  }
}

class _AttemptTile extends StatelessWidget {
  const _AttemptTile({required this.attempt});

  final QuizAttemptSummary attempt;

  @override
  Widget build(BuildContext context) {
    final pct = attempt.percentage.toStringAsFixed(0);
    final color = attempt.percentage >= 80
        ? Colors.green
        : attempt.percentage >= 60
            ? Colors.orange
            : Colors.red;
    return ListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(attempt.quizTitle),
      subtitle: Text(DateFormat('dd.MM.yyyy').format(attempt.finishedAt)),
      trailing: Text(
        '$pct%',
        style:
            TextStyle(color: color, fontWeight: FontWeight.bold, fontSize: 16),
      ),
    );
  }
}
