import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../features/ai/domain/entities/class_insights.dart';
import '../../../../features/ai/presentation/providers/ai_insights_provider.dart';
import '../../domain/entities/quiz.dart';

class QuizResultPage extends ConsumerWidget {
  const QuizResultPage({super.key, required this.result});

  final QuizResult result;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final pct = result.percentage;
    final color = pct >= 0.8
        ? Colors.green
        : pct >= 0.6
            ? Colors.orange
            : Colors.red;

    final gamification = ref.watch(gamificationProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Результат'),
        automaticallyImplyLeading: false,
      ),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          _ScoreHeader(result: result, color: color),
          const SizedBox(height: 12),
          gamification.when(
            data: (g) => _GamificationBanner(gamification: g),
            loading: () => const SizedBox.shrink(),
            error: (_, __) => const SizedBox.shrink(),
          ),
          const SizedBox(height: 16),
          Text(
            'Разбор ответов',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 12),
          ...result.questionResults.asMap().entries.map(
                (e) => _QuestionResultCard(
                  index: e.key + 1,
                  item: e.value,
                ),
              ),
          const SizedBox(height: 20),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Закрыть'),
          ),
        ],
      ),
    );
  }
}

class _GamificationBanner extends StatelessWidget {
  const _GamificationBanner({required this.gamification});

  final Gamification gamification;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      color: Colors.deepPurple.withValues(alpha: 0.07),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: Colors.deepPurple.withValues(alpha: 0.2)),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            _Stat(
              icon: Icons.bolt,
              iconColor: Colors.amber,
              value: '${gamification.xpTotal} XP',
              label: 'Всего опыта',
            ),
            Container(width: 1, height: 36, color: Colors.deepPurple.withValues(alpha: 0.15)),
            _Stat(
              icon: Icons.local_fire_department,
              iconColor: Colors.orange,
              value: '${gamification.streakDays} дн.',
              label: 'Серия',
            ),
            Container(width: 1, height: 36, color: Colors.deepPurple.withValues(alpha: 0.15)),
            _Stat(
              icon: Icons.emoji_events,
              iconColor: Colors.deepPurple,
              value: '${gamification.longestStreak} дн.',
              label: 'Рекорд',
            ),
          ],
        ),
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({
    required this.icon,
    required this.iconColor,
    required this.value,
    required this.label,
  });

  final IconData icon;
  final Color iconColor;
  final String value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, color: iconColor, size: 20),
        const SizedBox(height: 4),
        Text(value, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
        Text(label, style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
      ],
    );
  }
}

class _ScoreHeader extends StatelessWidget {
  const _ScoreHeader({required this.result, required this.color});

  final QuizResult result;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final score = result.attempt.score ?? 0;
    final max = result.attempt.maxScore;
    final pct = (result.percentage * 100).round();

    return Card(
      elevation: 0,
      color: color.withValues(alpha: 0.08),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: BorderSide(color: color.withValues(alpha: 0.3)),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 28, horizontal: 20),
        child: Column(
          children: [
            Text(
              '$pct%',
              style: Theme.of(context).textTheme.displaySmall?.copyWith(
                    color: color,
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 6),
            Text(
              'Правильных: $score / $max',
              style: Theme.of(context)
                  .textTheme
                  .titleMedium
                  ?.copyWith(color: color),
            ),
            const SizedBox(height: 16),
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              decoration: BoxDecoration(
                color: Colors.amber.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: Colors.amber.withValues(alpha: 0.4)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.star, size: 18, color: Colors.amber),
                  const SizedBox(width: 6),
                  Text(
                    '+${result.xpEarned} XP',
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Colors.amber,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _QuestionResultCard extends StatefulWidget {
  const _QuestionResultCard({required this.index, required this.item});

  final int index;
  final QuestionResult item;

  @override
  State<_QuestionResultCard> createState() => _QuestionResultCardState();
}

class _QuestionResultCardState extends State<_QuestionResultCard> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    final item = widget.item;
    final unanswered = item.selectedOptionId == null;
    final color = unanswered
        ? Colors.grey
        : item.isCorrect
            ? Colors.green
            : Colors.red;
    final icon = unanswered
        ? Icons.remove_circle_outline
        : item.isCorrect
            ? Icons.check_circle
            : Icons.cancel;

    return Card(
      margin: const EdgeInsets.only(bottom: 10),
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: color.withValues(alpha: 0.25)),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => setState(() => _expanded = !_expanded),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(icon, color: color, size: 22),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      '${widget.index}. ${item.questionBody}',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ),
                  Icon(
                    _expanded
                        ? Icons.expand_less
                        : Icons.expand_more,
                    color: Colors.grey,
                    size: 20,
                  ),
                ],
              ),
              if (_expanded) ...[
                const SizedBox(height: 12),
                const Divider(height: 1),
                const SizedBox(height: 10),
                if (!unanswered && !item.isCorrect) ...[
                  _AnswerRow(
                    label: 'Ваш ответ',
                    text: item.selectedBody ?? '',
                    color: Colors.red,
                    icon: Icons.close,
                  ),
                  const SizedBox(height: 6),
                ],
                _AnswerRow(
                  label: unanswered ? 'Правильный ответ' : 'Верный ответ',
                  text: item.correctBody,
                  color: Colors.green,
                  icon: Icons.check,
                ),
                if (unanswered) ...[
                  const SizedBox(height: 6),
                  Text(
                    'Вы не ответили на этот вопрос',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey.shade600,
                      fontStyle: FontStyle.italic,
                    ),
                  ),
                ],
                if (item.explanation != null &&
                    item.explanation!.isNotEmpty) ...[
                  const SizedBox(height: 10),
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: Colors.blue.withValues(alpha: 0.07),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Icon(
                          Icons.info_outline,
                          size: 16,
                          color: Colors.blue,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            item.explanation!,
                            style: TextStyle(
                              fontSize: 13,
                              color: Colors.blue.shade700,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _AnswerRow extends StatelessWidget {
  const _AnswerRow({
    required this.label,
    required this.text,
    required this.color,
    required this.icon,
  });

  final String label;
  final String text;
  final Color color;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 6),
        Expanded(
          child: RichText(
            text: TextSpan(
              style: DefaultTextStyle.of(context).style,
              children: [
                TextSpan(
                  text: '$label: ',
                  style: TextStyle(
                    fontSize: 13,
                    color: Colors.grey.shade600,
                  ),
                ),
                TextSpan(
                  text: text,
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                    color: color.withValues(alpha: 0.85),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
