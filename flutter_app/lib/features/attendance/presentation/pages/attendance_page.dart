import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/attendance.dart';
import '../providers/attendance_provider.dart';

class AttendancePage extends ConsumerWidget {
  const AttendancePage({
    super.key,
    required this.groupId,
    this.isTeacher = false,
    this.studentId,
  });

  final String groupId;
  final bool isTeacher;
  final String? studentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (isTeacher) {
      return _TeacherAttendancePage(groupId: groupId);
    }
    return _StudentAttendancePage(
      groupId: groupId,
      studentId: studentId ?? '',
    );
  }
}

// ─── Teacher view ─────────────────────────────────────────────────────────────

class _TeacherAttendancePage extends ConsumerWidget {
  const _TeacherAttendancePage({required this.groupId});

  final String groupId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(attendanceDateProvider(groupId));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Посещаемость'),
        actions: [
          IconButton(
            icon: const Icon(Icons.calendar_today),
            onPressed: () => _pickDate(context, ref, state.selectedDate),
          ),
        ],
      ),
      body: Column(
        children: [
          _DateBanner(date: state.selectedDate),
          if (state.isLoading)
            const Expanded(child: Center(child: CircularProgressIndicator()))
          else if (state.records.isEmpty)
            const Expanded(child: Center(child: Text('Нет записей на этот день')))
          else
            Expanded(
              child: ListView.builder(
                itemCount: state.records.length,
                itemBuilder: (_, i) => _AttendanceTile(
                  record: state.records[i],
                  onChangeStatus: (status) => ref
                      .read(attendanceDateProvider(groupId).notifier)
                      .mark(studentId: state.records[i].studentId, status: status),
                  onDelete: () => ref
                      .read(attendanceDateProvider(groupId).notifier)
                      .delete(state.records[i].id),
                ),
              ),
            ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _markSheet(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Отметить'),
      ),
    );
  }

  Future<void> _pickDate(
      BuildContext context, WidgetRef ref, DateTime? current) async {
    final picked = await showDatePicker(
      context: context,
      initialDate: current ?? DateTime.now(),
      firstDate: DateTime(2020),
      lastDate: DateTime.now(),
    );
    if (picked != null) {
      ref.read(attendanceDateProvider(groupId).notifier).load(picked);
    }
  }

  void _markSheet(BuildContext context, WidgetRef ref) {
    final studentIdCtrl = TextEditingController();
    AttendanceStatus selected = AttendanceStatus.present;

    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setState) {
          final insets = MediaQuery.viewInsetsOf(ctx);
          return Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 16 + insets.bottom),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text('Отметить посещаемость',
                    style: Theme.of(ctx).textTheme.titleLarge),
                const SizedBox(height: 16),
                TextField(
                  controller: studentIdCtrl,
                  decoration: const InputDecoration(labelText: 'ID студента'),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<AttendanceStatus>(
                  value: selected,
                  decoration: const InputDecoration(labelText: 'Статус'),
                  items: AttendanceStatus.values
                      .map((s) => DropdownMenuItem(
                            value: s,
                            child: Text(s.label),
                          ))
                      .toList(),
                  onChanged: (v) => setState(() => selected = v!),
                ),
                const SizedBox(height: 20),
                FilledButton(
                  onPressed: () {
                    ref
                        .read(attendanceDateProvider(groupId).notifier)
                        .mark(studentId: studentIdCtrl.text.trim(), status: selected);
                    Navigator.of(ctx).pop();
                  },
                  child: const Text('Сохранить'),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}

// ─── Student view ─────────────────────────────────────────────────────────────

class _StudentAttendancePage extends ConsumerWidget {
  const _StudentAttendancePage(
      {required this.groupId, required this.studentId});

  final String groupId;
  final String studentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(
      studentAttendanceProvider((groupId: groupId, studentId: studentId)),
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Моя посещаемость')),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Ошибка: $e')),
        data: (list) {
          if (list.isEmpty) return const Center(child: Text('Нет данных'));
          final stats = _stats(list);
          return Column(
            children: [
              _StatsBanner(stats: stats),
              Expanded(
                child: ListView.builder(
                  itemCount: list.length,
                  itemBuilder: (_, i) =>
                      _AttendanceTile(record: list[i]),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Map<AttendanceStatus, int> _stats(List<Attendance> list) {
    final m = <AttendanceStatus, int>{};
    for (final a in list) {
      m[a.status] = (m[a.status] ?? 0) + 1;
    }
    return m;
  }
}

// ─── Date banner ──────────────────────────────────────────────────────────────

class _DateBanner extends StatelessWidget {
  const _DateBanner({required this.date});

  final DateTime? date;

  @override
  Widget build(BuildContext context) {
    if (date == null) return const SizedBox.shrink();
    final d = date!;
    final label =
        '${d.day.toString().padLeft(2, '0')}.${d.month.toString().padLeft(2, '0')}.${d.year}';
    return Container(
      color: Theme.of(context).colorScheme.primaryContainer,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          const Icon(Icons.calendar_today, size: 16),
          const SizedBox(width: 8),
          Text(label, style: Theme.of(context).textTheme.bodyMedium),
        ],
      ),
    );
  }
}

// ─── Stats banner ─────────────────────────────────────────────────────────────

class _StatsBanner extends StatelessWidget {
  const _StatsBanner({required this.stats});

  final Map<AttendanceStatus, int> stats;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Theme.of(context).colorScheme.surfaceVariant,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: AttendanceStatus.values.map((s) {
          return Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('${stats[s] ?? 0}',
                  style: Theme.of(context).textTheme.titleMedium),
              Text(s.label,
                  style: Theme.of(context).textTheme.labelSmall),
            ],
          );
        }).toList(),
      ),
    );
  }
}

// ─── Attendance tile ──────────────────────────────────────────────────────────

class _AttendanceTile extends StatelessWidget {
  const _AttendanceTile({
    required this.record,
    this.onChangeStatus,
    this.onDelete,
  });

  final Attendance record;
  final void Function(AttendanceStatus)? onChangeStatus;
  final VoidCallback? onDelete;

  Color _statusColor(AttendanceStatus s) => switch (s) {
        AttendanceStatus.present => Colors.green,
        AttendanceStatus.absent => Colors.red,
        AttendanceStatus.late => Colors.orange,
        AttendanceStatus.excused => Colors.blue,
      };

  @override
  Widget build(BuildContext context) {
    final color = _statusColor(record.status);
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: color.withOpacity(0.15),
        child: Icon(_statusIcon(record.status), color: color, size: 20),
      ),
      title: Text(record.studentId),
      subtitle: Text('${record.lessonDate}  •  ${record.status.label}'),
      trailing: onDelete != null
          ? Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (onChangeStatus != null)
                  PopupMenuButton<AttendanceStatus>(
                    icon: const Icon(Icons.edit_outlined),
                    onSelected: onChangeStatus,
                    itemBuilder: (_) => AttendanceStatus.values
                        .map((s) => PopupMenuItem(
                              value: s,
                              child: Text(s.label),
                            ))
                        .toList(),
                  ),
                IconButton(
                  icon: const Icon(Icons.delete_outline, color: Colors.red),
                  onPressed: onDelete,
                ),
              ],
            )
          : null,
    );
  }

  IconData _statusIcon(AttendanceStatus s) => switch (s) {
        AttendanceStatus.present => Icons.check_circle,
        AttendanceStatus.absent => Icons.cancel,
        AttendanceStatus.late => Icons.watch_later,
        AttendanceStatus.excused => Icons.info,
      };
}
