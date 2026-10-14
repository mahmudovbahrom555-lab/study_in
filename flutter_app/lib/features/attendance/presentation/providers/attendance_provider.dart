import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/network/error_message.dart';
import '../../data/attendance_api.dart';
import '../../data/attendance_repository_impl.dart';
import '../../domain/entities/attendance.dart';
import '../../domain/repositories/attendance_repository.dart';

final attendanceRepositoryProvider = Provider<AttendanceRepository>((ref) {
  return AttendanceRepositoryImpl(AttendanceApi(ref.watch(dioProvider)));
});

// ─── Список занятия (репетитор) ───────────────────────────────────────────────

class RosterState {
  const RosterState({
    required this.date,
    this.entries = const [],
    this.saving = const {},
    this.isLoading = false,
    this.loadError,
    this.saveError,
  });

  final DateTime date;
  final List<RosterEntry> entries;

  /// Ученики, чья отметка сейчас сохраняется: повторный тап игнорируем.
  final Set<String> saving;
  final bool isLoading;
  final Object? loadError;

  /// Ошибка сохранения — показывается разово (SnackBar), отметка откатывается.
  final String? saveError;

  int get marked => entries.where((e) => e.status != null).length;
  bool get hasUnmarked => entries.any((e) => e.status == null);

  RosterState copyWith({
    DateTime? date,
    List<RosterEntry>? entries,
    Set<String>? saving,
    bool? isLoading,
    Object? loadError,
    bool clearLoadError = false,
    String? saveError,
  }) =>
      RosterState(
        date: date ?? this.date,
        entries: entries ?? this.entries,
        saving: saving ?? this.saving,
        isLoading: isLoading ?? this.isLoading,
        loadError: clearLoadError ? null : (loadError ?? this.loadError),
        saveError: saveError,
      );
}

class RosterNotifier extends StateNotifier<RosterState> {
  RosterNotifier(this._repo, this._groupId)
      : super(RosterState(date: _today())) {
    load(state.date);
  }

  final AttendanceRepository _repo;
  final String _groupId;

  static DateTime _today() {
    final n = DateTime.now();
    return DateTime(n.year, n.month, n.day);
  }

  Future<void> load(DateTime date) async {
    final day = DateTime(date.year, date.month, date.day);
    state = state.copyWith(date: day, isLoading: true, clearLoadError: true);
    try {
      final entries = await _repo.roster(_groupId, _fmt(day));
      // Пока грузили, могли переключить день — старый ответ не нужен.
      if (state.date != day) return;
      state = state.copyWith(entries: entries, isLoading: false);
    } catch (e) {
      if (state.date != day) return;
      state = state.copyWith(isLoading: false, loadError: e);
    }
  }

  void shiftDay(int days) {
    final next = state.date.add(Duration(days: days));
    if (next.isAfter(_today())) return;
    load(next);
  }

  /// Тап по статусу: ставит его, а повторный тап по тому же статусу снимает
  /// отметку. Экран меняется сразу; при ошибке отметка откатывается.
  Future<void> toggle(String studentId, AttendanceStatus status) async {
    if (state.saving.contains(studentId)) return;
    final before = _entry(studentId);
    if (before == null) return;
    final clearing = before.status == status;
    final day = state.date;

    _replace(before.withMark(before.attendanceId, clearing ? null : status));
    state = state.copyWith(saving: {...state.saving, studentId});
    String? error;
    try {
      if (clearing) {
        if (before.attendanceId != null) {
          await _repo.delete(_groupId, before.attendanceId!);
        }
        _apply(day, studentId, (e) => e.withMark(null, null));
      } else {
        final a = await _repo.mark(
          groupId: _groupId,
          studentId: studentId,
          lessonDate: _fmt(day),
          status: status,
        );
        _apply(day, studentId, (e) => e.withMark(a.id, a.status));
      }
    } catch (e) {
      _apply(day, studentId, (_) => before);
      if (state.date == day) error = userErrorMessage(e);
    } finally {
      state = state.copyWith(
        saving: {...state.saving}..remove(studentId),
        saveError: error,
      );
    }
  }

  /// «Остальные были»: всем неотмеченным — «присутствовал», одним запросом.
  Future<void> markRestPresent() async {
    final rest = state.entries
        .where((e) => e.status == null && !state.saving.contains(e.studentId))
        .toList();
    if (rest.isEmpty) return;
    final day = state.date;
    final ids = {for (final e in rest) e.studentId};

    state = state.copyWith(
      entries: [
        for (final e in state.entries)
          ids.contains(e.studentId)
              ? e.withMark(null, AttendanceStatus.present)
              : e,
      ],
      saving: {...state.saving, ...ids},
    );
    String? error;
    try {
      final saved = await _repo.markMany(
        groupId: _groupId,
        lessonDate: _fmt(day),
        marks: {for (final id in ids) id: AttendanceStatus.present},
      );
      for (final a in saved) {
        _apply(day, a.studentId, (e) => e.withMark(a.id, a.status));
      }
    } catch (e) {
      for (final r in rest) {
        _apply(day, r.studentId, (_) => r);
      }
      if (state.date == day) error = userErrorMessage(e);
    } finally {
      state = state.copyWith(
        saving: {...state.saving}..removeAll(ids),
        saveError: error,
      );
    }
  }

  RosterEntry? _entry(String studentId) =>
      state.entries.where((e) => e.studentId == studentId).firstOrNull;

  void _replace(RosterEntry entry) {
    state = state.copyWith(
      entries: [
        for (final e in state.entries)
          e.studentId == entry.studentId ? entry : e,
      ],
    );
  }

  /// Применяет результат сохранения, только если на экране всё ещё тот же день.
  void _apply(DateTime day, String studentId, RosterEntry Function(RosterEntry) f) {
    if (state.date != day) return;
    final e = _entry(studentId);
    if (e != null) _replace(f(e));
  }

  static String _fmt(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-'
      '${d.month.toString().padLeft(2, '0')}-'
      '${d.day.toString().padLeft(2, '0')}';
}

final rosterProvider = StateNotifierProvider.autoDispose
    .family<RosterNotifier, RosterState, String>((ref, groupId) {
  return RosterNotifier(ref.watch(attendanceRepositoryProvider), groupId);
});

// ─── История ученика ──────────────────────────────────────────────────────────

final studentAttendanceProvider = StateNotifierProvider.autoDispose.family<
    _StudentAttendanceNotifier,
    AsyncValue<List<Attendance>>,
    ({String groupId, String studentId})>((ref, params) {
  return _StudentAttendanceNotifier(
    ref.watch(attendanceRepositoryProvider),
    params.groupId,
    params.studentId,
  );
});

class _StudentAttendanceNotifier
    extends StateNotifier<AsyncValue<List<Attendance>>> {
  _StudentAttendanceNotifier(this._repo, this._groupId, this._studentId)
      : super(const AsyncValue.loading()) {
    _load();
  }

  final AttendanceRepository _repo;
  final String _groupId;
  final String _studentId;

  Future<void> _load() async {
    state = const AsyncValue.loading();
    try {
      final list = await _repo.listStudent(_groupId, _studentId);
      state = AsyncValue.data(list);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}
