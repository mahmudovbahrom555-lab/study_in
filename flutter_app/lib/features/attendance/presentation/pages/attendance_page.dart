import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/theme/colors.dart';
import '../../../../core/widgets/error_view.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/attendance.dart';
import '../attendance_status_l10n.dart';
import '../providers/attendance_provider.dart';

class AttendancePage extends ConsumerWidget {
  const AttendancePage({
    super.key,
    required this.groupId,
    this.isTeacher = false,
  });

  final String groupId;
  final bool isTeacher;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (isTeacher) return _TeacherAttendancePage(groupId: groupId);
    final params = (
      groupId: groupId,
      studentId: ref.watch(authProvider).user?.id ?? '',
    );
    return AttendanceHistoryView(
      title: context.l10n.myAttendance,
      records: ref.watch(studentAttendanceProvider(params)),
      onRetry: () => ref.invalidate(studentAttendanceProvider(params)),
    );
  }
}

Color _statusColor(AttendanceStatus s) => switch (s) {
      AttendanceStatus.present => AppColors.success,
      AttendanceStatus.absent => AppColors.error,
      AttendanceStatus.late => AppColors.warning,
      AttendanceStatus.excused => AppColors.info,
    };

IconData _statusIcon(AttendanceStatus s) => switch (s) {
      AttendanceStatus.present => LucideIcons.circleCheck,
      AttendanceStatus.absent => LucideIcons.circleX,
      AttendanceStatus.late => LucideIcons.clock,
      AttendanceStatus.excused => LucideIcons.info,
    };

String _fmtDate(DateTime d) => DateFormat('dd.MM.yyyy').format(d);

// ─── Репетитор: список занятия ────────────────────────────────────────────────

class _TeacherAttendancePage extends ConsumerWidget {
  const _TeacherAttendancePage({required this.groupId});

  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = context.l10n;
    final state = ref.watch(rosterProvider(groupId));
    final notifier = ref.read(rosterProvider(groupId).notifier);

    ref.listen(rosterProvider(groupId).select((s) => s.saveError), (_, err) {
      if (err != null) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(err)));
      }
    });

    final Widget body;
    if (state.loadError != null && state.entries.isEmpty) {
      body = ErrorView(error: state.loadError, onRetry: () => notifier.load(state.date));
    } else if (state.isLoading && state.entries.isEmpty) {
      body = const Center(child: CircularProgressIndicator());
    } else if (state.entries.isEmpty) {
      body = Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Text(l10n.attNoStudents, textAlign: TextAlign.center),
        ),
      );
    } else {
      body = RefreshIndicator(
        onRefresh: () => notifier.load(state.date),
        child: ListView.separated(
          padding: const EdgeInsets.only(bottom: 24),
          itemCount: state.entries.length,
          separatorBuilder: (_, __) => const Divider(height: 1, indent: 16),
          itemBuilder: (_, i) {
            final e = state.entries[i];
            return _RosterRow(
              entry: e,
              saving: state.saving.contains(e.studentId),
              onTap: (s) => notifier.toggle(e.studentId, s),
            );
          },
        ),
      );
    }

    return Scaffold(
      appBar: AppBar(title: Text(l10n.attendance)),
      body: Column(
        children: [
          _DayBar(groupId: groupId),
          if (state.entries.isNotEmpty) _Summary(state: state, onRestPresent: notifier.markRestPresent),
          SizedBox(
            height: 2,
            child: state.isLoading && state.entries.isNotEmpty
                ? const LinearProgressIndicator(minHeight: 2)
                : null,
          ),
          Expanded(child: body),
        ],
      ),
    );
  }
}

class _DayBar extends ConsumerWidget {
  const _DayBar({required this.groupId});

  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = context.l10n;
    final date = ref.watch(rosterProvider(groupId).select((s) => s.date));
    final notifier = ref.read(rosterProvider(groupId).notifier);
    final now = DateTime.now();
    final isToday = date == DateTime(now.year, now.month, now.day);

    return Material(
      color: Theme.of(context).colorScheme.surface,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Row(
          children: [
            IconButton(
              tooltip: l10n.attPrevDay,
              icon: const Icon(LucideIcons.chevronLeft),
              onPressed: () => notifier.shiftDay(-1),
            ),
            Expanded(
              child: TextButton.icon(
                icon: const Icon(LucideIcons.calendar, size: 18),
                label: Text(isToday ? l10n.attToday(_fmtDate(date)) : _fmtDate(date)),
                onPressed: () async {
                  final picked = await showDatePicker(
                    context: context,
                    initialDate: date,
                    firstDate: DateTime(2020),
                    lastDate: DateTime.now(),
                    helpText: l10n.pickDate,
                  );
                  if (picked != null) notifier.load(picked);
                },
              ),
            ),
            IconButton(
              tooltip: l10n.attNextDay,
              icon: const Icon(LucideIcons.chevronRight),
              // Будущие дни не отмечаем.
              onPressed: isToday ? null : () => notifier.shiftDay(1),
            ),
          ],
        ),
      ),
    );
  }
}

class _Summary extends StatelessWidget {
  const _Summary({required this.state, required this.onRestPresent});

  final RosterState state;
  final VoidCallback onRestPresent;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    // Высота постоянная: когда кнопка «Остальные…» исчезает, список не должен
    // прыгать — иначе следующий тап попадёт по соседнему ученику.
    return Container(
      height: 56,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: [
          Expanded(
            child: Text(
              l10n.attMarkedOf(state.marked, state.entries.length),
              style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant),
            ),
          ),
          if (state.hasUnmarked)
            FilledButton.tonalIcon(
              icon: const Icon(LucideIcons.checkCheck, size: 18),
              label: Text(l10n.attRestPresent),
              onPressed: onRestPresent,
            ),
        ],
      ),
    );
  }
}

/// Строка ученика: имя и статус текстом, справа 4 кнопки статуса.
/// Тап ставит статус, повторный тап по нему же снимает отметку.
class _RosterRow extends StatelessWidget {
  const _RosterRow({required this.entry, required this.saving, required this.onTap});

  final RosterEntry entry;
  final bool saving;
  final void Function(AttendanceStatus) onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = context.l10n;
    final status = entry.status;
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 6, 8, 6),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  entry.studentName.isEmpty ? '—' : entry.studentName,
                  style: theme.textTheme.titleMedium,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                Text(
                  status?.label(l10n) ?? l10n.attNotMarked,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: status == null ? theme.colorScheme.onSurfaceVariant : _statusColor(status),
                  ),
                ),
              ],
            ),
          ),
          for (final s in AttendanceStatus.values)
            _StatusButton(
              status: s,
              selected: status == s,
              enabled: !saving,
              onTap: () => onTap(s),
            ),
        ],
      ),
    );
  }
}

class _StatusButton extends StatelessWidget {
  const _StatusButton({
    required this.status,
    required this.selected,
    required this.enabled,
    required this.onTap,
  });

  final AttendanceStatus status;
  final bool selected;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final color = _statusColor(status);
    final muted = Theme.of(context).colorScheme.outline;
    return Semantics(
      selected: selected,
      child: IconButton(
        tooltip: status.label(context.l10n),
        isSelected: selected,
        onPressed: enabled ? onTap : null,
        style: IconButton.styleFrom(
          backgroundColor: selected ? color.withValues(alpha: 0.15) : null,
        ),
        icon: Icon(_statusIcon(status), color: selected ? color : muted),
      ),
    );
  }
}

// ─── Ученик/родитель: история ─────────────────────────────────────────────────

/// История посещаемости одного ученика. Данные приходят снаружи: ученик берёт
/// свою, родитель — ребёнка через свой модуль (разные эндпоинты и права).
class AttendanceHistoryView extends StatelessWidget {
  const AttendanceHistoryView({
    super.key,
    required this.title,
    required this.records,
    required this.onRetry,
  });

  final String title;
  final AsyncValue<List<Attendance>> records;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: records.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => ErrorView(error: e, onRetry: onRetry),
        data: (list) {
          if (list.isEmpty) return Center(child: Text(context.l10n.noData));
          return Column(
            children: [
              _StatsBanner(list: list),
              Expanded(
                child: ListView.builder(
                  itemCount: list.length,
                  itemBuilder: (_, i) => _HistoryTile(record: list[i]),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _StatsBanner extends StatelessWidget {
  const _StatsBanner({required this.list});

  final List<Attendance> list;

  @override
  Widget build(BuildContext context) {
    final counts = <AttendanceStatus, int>{};
    for (final a in list) {
      counts[a.status] = (counts[a.status] ?? 0) + 1;
    }
    return Container(
      color: Theme.of(context).colorScheme.surfaceContainerHighest,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          for (final s in AttendanceStatus.values)
            Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text('${counts[s] ?? 0}', style: Theme.of(context).textTheme.titleMedium),
                Text(s.label(context.l10n), style: Theme.of(context).textTheme.labelSmall),
              ],
            ),
        ],
      ),
    );
  }
}

class _HistoryTile extends StatelessWidget {
  const _HistoryTile({required this.record});

  final Attendance record;

  @override
  Widget build(BuildContext context) {
    final color = _statusColor(record.status);
    final date = DateTime.tryParse(record.lessonDate);
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: color.withValues(alpha: 0.15),
        child: Icon(_statusIcon(record.status), color: color, size: 20),
      ),
      title: Text(record.status.label(context.l10n)),
      subtitle: Text(date == null ? record.lessonDate : _fmtDate(date)),
    );
  }
}
