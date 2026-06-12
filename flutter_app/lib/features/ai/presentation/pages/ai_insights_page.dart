import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/router/routes.dart';
import '../../domain/entities/class_insights.dart';
import '../providers/ai_insights_provider.dart';

class AiInsightsPage extends ConsumerWidget {
  const AiInsightsPage({super.key, required this.groupId});

  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(classInsightsProvider(groupId));

    return Scaffold(
      appBar: AppBar(title: const Text('AI Insights')),
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (insights) => _InsightsBody(groupId: groupId, insights: insights),
      ),
    );
  }
}

class _InsightsBody extends StatelessWidget {
  const _InsightsBody({required this.groupId, required this.insights});

  final String groupId;
  final ClassInsights insights;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        _StatRow(insights: insights),
        if (insights.recommendations.isNotEmpty) ...[
          const SizedBox(height: 20),
          _NextBestActionPanel(recommendations: insights.recommendations, groupId: groupId),
        ],
        const SizedBox(height: 20),
        Text('Слабые темы класса', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        if (insights.classWeakness.isEmpty)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 8),
            child: Text('Нет данных — тесты ещё не проходились'),
          )
        else
          ...insights.classWeakness.map((t) => _WeakTopicTile(topic: t)),
        const SizedBox(height: 20),
        Text('Таблица лидеров', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        ...insights.students.asMap().entries.map(
              (e) => _StudentTile(
                rank: e.key + 1,
                student: e.value,
                groupId: groupId,
              ),
            ),
      ],
    );
  }
}

// ─── Next Best Action ─────────────────────────────────────────────────────────

class _NextBestActionPanel extends StatelessWidget {
  const _NextBestActionPanel({required this.recommendations, required this.groupId});

  final List<TeacherRecommendation> recommendations;
  final String groupId;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Рекомендации', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        ...recommendations.map((r) => _RecommendationCard(rec: r, groupId: groupId)),
      ],
    );
  }
}

class _RecommendationCard extends ConsumerStatefulWidget {
  const _RecommendationCard({required this.rec, required this.groupId});

  final TeacherRecommendation rec;
  final String groupId;

  @override
  ConsumerState<_RecommendationCard> createState() => _RecommendationCardState();
}

class _RecommendationCardState extends ConsumerState<_RecommendationCard> {
  bool _dismissed = false;
  bool _accepted = false;

  Future<void> _act(String status) async {
    await ref
        .read(aiRepositoryProvider)
        .recordRecommendationAction(widget.rec.id, status: status);
    if (!mounted) return;
    setState(() {
      _dismissed = status == 'dismissed';
      _accepted = status == 'accepted';
    });
  }

  Future<void> _explain() async {
    final text = await ref.read(aiRepositoryProvider).explainRecommendation(widget.rec.id);
    if (!mounted) return;
    showModalBottomSheet<void>(
      context: context,
      builder: (_) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Объяснение AI', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 12),
            Text(text),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_dismissed) return const SizedBox.shrink();

    final theme = Theme.of(context);
    final (icon, color) = _iconFor(widget.rec.action);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(10),
        side: BorderSide(
          color: _accepted ? Colors.green.withValues(alpha: 0.6) : color.withValues(alpha: 0.4),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ListTile(
            leading: CircleAvatar(
              backgroundColor: color.withValues(alpha: 0.12),
              child: Icon(icon, color: color, size: 20),
            ),
            title: Text(
              widget.rec.topic != null
                  ? '${_labelFor(widget.rec.action)}: ${widget.rec.topic}'
                  : _labelFor(widget.rec.action),
              style: theme.textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
            ),
            subtitle: Text(widget.rec.reason, style: theme.textTheme.bodySmall),
            trailing: _PriorityBadge(priority: widget.rec.priority),
          ),
          if (!_accepted)
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 0, 12, 8),
              child: Row(
                children: [
                  TextButton.icon(
                    onPressed: () => _act('accepted'),
                    icon: const Icon(Icons.check, size: 16),
                    label: const Text('Принять'),
                    style: TextButton.styleFrom(foregroundColor: Colors.green),
                  ),
                  TextButton.icon(
                    onPressed: () => _act('dismissed'),
                    icon: const Icon(Icons.close, size: 16),
                    label: const Text('Отклонить'),
                    style: TextButton.styleFrom(foregroundColor: Colors.grey),
                  ),
                  const Spacer(),
                  TextButton.icon(
                    onPressed: _explain,
                    icon: const Icon(Icons.auto_awesome, size: 16),
                    label: const Text('Почему?'),
                    style: TextButton.styleFrom(foregroundColor: Colors.indigo),
                  ),
                ],
              ),
            )
          else
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 10),
              child: Row(
                children: [
                  const Icon(Icons.check_circle, color: Colors.green, size: 16),
                  const SizedBox(width: 4),
                  Text('Принято', style: theme.textTheme.bodySmall?.copyWith(color: Colors.green)),
                ],
              ),
            ),
        ],
      ),
    );
  }

  (IconData, Color) _iconFor(String action) => switch (action) {
        'create_quiz' => (Icons.quiz_outlined, Colors.blue),
        'schedule_review' => (Icons.event_repeat, Colors.orange),
        'check_students' => (Icons.person_search, Colors.red),
        'rate_questions' => (Icons.rate_review_outlined, Colors.purple),
        'celebrate' => (Icons.celebration, Colors.green),
        _ => (Icons.lightbulb_outline, Colors.grey),
      };

  String _labelFor(String action) => switch (action) {
        'create_quiz' => 'Создать тест',
        'schedule_review' => 'Назначить повторение',
        'check_students' => 'Написать ученикам',
        'rate_questions' => 'Оценить вопросы',
        'celebrate' => 'Отличный результат',
        _ => 'Рекомендация',
      };
}

class _PriorityBadge extends StatelessWidget {
  const _PriorityBadge({required this.priority});

  final int priority;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (priority) {
      1 => ('Важно', Colors.red),
      2 => ('Средне', Colors.orange),
      _ => ('Инфо', Colors.grey),
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha:0.12),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(label, style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.w600)),
    );
  }
}

class _StatRow extends StatelessWidget {
  const _StatRow({required this.insights});

  final ClassInsights insights;

  @override
  Widget build(BuildContext context) {
    final pct = (insights.quizStats.acceptanceRate * 100).toStringAsFixed(1);
    return Row(
      children: [
        _StatCard(
          label: 'Учеников',
          value: '${insights.studentCount}',
          icon: Icons.group,
        ),
        const SizedBox(width: 8),
        _StatCard(
          label: 'TAR',
          value: '$pct%',
          icon: Icons.thumb_up_outlined,
          color: _tarColor(insights.quizStats.acceptanceRate),
        ),
        const SizedBox(width: 8),
        _StatCard(
          label: 'Средний балл',
          value: '${insights.quizStats.avgScore.toStringAsFixed(1)}%',
          icon: Icons.bar_chart,
        ),
      ],
    );
  }

  Color _tarColor(double rate) {
    if (rate >= 0.8) return Colors.green;
    if (rate >= 0.6) return Colors.orange;
    return Colors.red;
  }
}

class _StatCard extends StatelessWidget {
  const _StatCard({
    required this.label,
    required this.value,
    required this.icon,
    this.color,
  });

  final String label;
  final String value;
  final IconData icon;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Expanded(
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            children: [
              Icon(icon, color: color ?? theme.colorScheme.primary),
              const SizedBox(height: 4),
              Text(value, style: theme.textTheme.titleLarge?.copyWith(color: color)),
              Text(label, style: theme.textTheme.bodySmall),
            ],
          ),
        ),
      ),
    );
  }
}

class _WeakTopicTile extends StatelessWidget {
  const _WeakTopicTile({required this.topic});

  final TopicWeakness topic;

  @override
  Widget build(BuildContext context) {
    final pct = (topic.avgAccuracy * 100).toStringAsFixed(0);
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: CircleAvatar(
        backgroundColor: Colors.red.shade50,
        child: Text('$pct%', style: const TextStyle(fontSize: 12, color: Colors.red)),
      ),
      title: Text(topic.topic),
      subtitle: Text('${topic.studentsStruggling}/${topic.totalStudents} учеников'),
      trailing: SizedBox(
        width: 60,
        child: LinearProgressIndicator(
          value: topic.avgAccuracy,
          color: Colors.red,
          backgroundColor: Colors.red.shade100,
        ),
      ),
    );
  }
}

class _StudentTile extends StatelessWidget {
  const _StudentTile({required this.rank, required this.student, required this.groupId});

  final int rank;
  final StudentSummary student;
  final String groupId;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: _RankBadge(rank: rank),
      title: Row(
        children: [
          Expanded(child: Text(student.name)),
          if (student.isAtRisk)
            const Icon(Icons.warning_amber_rounded, color: Colors.orange, size: 18),
        ],
      ),
      subtitle: Text(
        '${student.xpTotal} XP · 🔥${student.streakDays} дней'
        '${student.topWeakness.isNotEmpty ? ' · слабо: ${student.topWeakness}' : ''}',
      ),
      trailing: const Icon(Icons.chevron_right),
      onTap: () => context.push(Routes.studentProgress(groupId, student.studentId)),
    );
  }
}

class _RankBadge extends StatelessWidget {
  const _RankBadge({required this.rank});

  final int rank;

  @override
  Widget build(BuildContext context) {
    final color = switch (rank) {
      1 => Colors.amber,
      2 => Colors.blueGrey.shade300,
      3 => Colors.brown.shade300,
      _ => Colors.grey.shade300,
    };
    return CircleAvatar(
      backgroundColor: color,
      child: Text('$rank', style: const TextStyle(fontWeight: FontWeight.bold)),
    );
  }
}

