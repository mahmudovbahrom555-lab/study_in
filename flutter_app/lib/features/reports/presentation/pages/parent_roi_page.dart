import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/report.dart';
import '../providers/reports_provider.dart';

class ParentRoiPage extends ConsumerWidget {
  const ParentRoiPage({
    super.key,
    required this.studentId,
    required this.groupId,
    this.studentName,
  });

  final String studentId;
  final String groupId;
  final String? studentName;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(
      parentRoiProvider((studentId: studentId, groupId: groupId)),
    );

    return Scaffold(
      appBar: AppBar(
        title: Text(studentName != null ? 'Прогресс — $studentName' : 'Прогресс'),
      ),
      body: async.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (report) => _ReportBody(report: report),
      ),
    );
  }
}

class _ReportBody extends StatelessWidget {
  const _ReportBody({required this.report});

  final ParentReport report;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          'Период: ${report.periodStart} — ${report.periodEnd}',
          style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey),
        ),
        const SizedBox(height: 16),
        if (report.aiSummary != null) ...[
          _SummaryCard(summary: report.aiSummary!),
          const SizedBox(height: 20),
        ],
        Text('Ключевые показатели', style: theme.textTheme.titleMedium),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _MetricCard(
                title: 'Посещаемость',
                value: '${report.attendancePct.toStringAsFixed(0)}%',
                color: _attendanceColor(report.attendancePct),
                icon: Icons.calendar_today_outlined,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _MetricCard(
                title: 'Тесты',
                value: '${report.quizScoreAvg.toStringAsFixed(0)}%',
                delta: report.quizScoreDelta,
                icon: Icons.quiz_outlined,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _MetricCard(
                title: 'Освоение',
                value: '${(report.masteryAvg * 100).toStringAsFixed(0)}%',
                delta: report.masteryDelta,
                icon: Icons.auto_graph_outlined,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _MetricCard(
                title: 'Домашние задания',
                value: '${report.homeworkCompletionPct.toStringAsFixed(0)}%',
                icon: Icons.assignment_turned_in_outlined,
                color: Colors.teal,
              ),
            ),
          ],
        ),
        const SizedBox(height: 20),
        _StatsRow(report: report),
      ],
    );
  }

  Color _attendanceColor(double pct) {
    if (pct >= 90) return Colors.green;
    if (pct >= 75) return Colors.orange;
    return Colors.red;
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard({required this.summary});

  final String summary;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.blue.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.blue.shade100),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.auto_awesome, color: Colors.blue.shade700, size: 18),
              const SizedBox(width: 8),
              Text(
                'Резюме',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: Colors.blue.shade700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(summary, style: const TextStyle(fontSize: 15, height: 1.5)),
        ],
      ),
    );
  }
}

class _MetricCard extends StatelessWidget {
  const _MetricCard({
    required this.title,
    required this.value,
    required this.icon,
    this.delta,
    this.color,
  });

  final String title;
  final String value;
  final IconData icon;
  final double? delta;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final effectiveColor = color ??
        (delta == null
            ? Theme.of(context).colorScheme.primary
            : delta! >= 0
                ? Colors.green
                : Colors.red);

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: effectiveColor.withValues(alpha: 0.07),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: effectiveColor.withValues(alpha: 0.25)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 16, color: effectiveColor),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  title,
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            value,
            style: TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.bold,
              color: effectiveColor,
            ),
          ),
          if (delta != null) ...[
            const SizedBox(height: 4),
            Text(
              '${delta! >= 0 ? '+' : ''}${delta!.toStringAsFixed(1)}% vs пр. мес.',
              style: TextStyle(
                fontSize: 11,
                color: delta! >= 0 ? Colors.green : Colors.red,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _StatsRow extends StatelessWidget {
  const _StatsRow({required this.report});

  final ParentReport report;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: [
          _Mini(
            label: 'Попыток\nтестов',
            value: '${report.quizAttemptsCount}',
          ),
          Container(width: 1, height: 32, color: Colors.grey.shade300),
          _Mini(
            label: 'Балл\n(тесты)',
            value: '${report.quizScoreAvg.toStringAsFixed(0)}%',
          ),
          Container(width: 1, height: 32, color: Colors.grey.shade300),
          _Mini(
            label: 'ДЗ\nсдано',
            value: '${report.homeworkCompletionPct.toStringAsFixed(0)}%',
          ),
        ],
      ),
    );
  }
}

class _Mini extends StatelessWidget {
  const _Mini({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          value,
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          textAlign: TextAlign.center,
          style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
        ),
      ],
    );
  }
}
