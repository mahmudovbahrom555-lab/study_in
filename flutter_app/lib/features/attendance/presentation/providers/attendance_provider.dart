import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/attendance_api.dart';
import '../../data/attendance_repository_impl.dart';
import '../../domain/entities/attendance.dart';
import '../../domain/repositories/attendance_repository.dart';

final attendanceRepositoryProvider =
    Provider.family<AttendanceRepository, String>((ref, groupId) {
  final dio = ref.watch(dioProvider);
  return AttendanceRepositoryImpl(AttendanceApi(dio));
});

// ─── Attendance by date (teacher) ─────────────────────────────────────────────

class AttendanceDateState {
  const AttendanceDateState({
    this.records = const [],
    this.selectedDate,
    this.isLoading = false,
    this.error,
  });

  final List<Attendance> records;
  final DateTime? selectedDate;
  final bool isLoading;
  final String? error;

  AttendanceDateState copyWith({
    List<Attendance>? records,
    DateTime? selectedDate,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      AttendanceDateState(
        records: records ?? this.records,
        selectedDate: selectedDate ?? this.selectedDate,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

class AttendanceDateNotifier extends StateNotifier<AttendanceDateState> {
  AttendanceDateNotifier(this._repo, this._groupId)
      : super(AttendanceDateState(selectedDate: DateTime.now())) {
    load(DateTime.now());
  }

  final AttendanceRepository _repo;
  final String _groupId;

  Future<void> load(DateTime date) async {
    final dateStr = _fmt(date);
    state = state.copyWith(isLoading: true, selectedDate: date, clearError: true);
    try {
      final list = await _repo.listByDate(_groupId, dateStr);
      state = state.copyWith(isLoading: false, records: list);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> mark({
    required String studentId,
    required AttendanceStatus status,
    String? note,
  }) async {
    final date = state.selectedDate ?? DateTime.now();
    try {
      final a = await _repo.mark(
        groupId: _groupId,
        studentId: studentId,
        lessonDate: _fmt(date),
        status: status,
        note: note,
      );
      final updated = [
        a,
        ...state.records.where((r) => r.studentId != studentId),
      ];
      state = state.copyWith(records: updated);
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> delete(String id) async {
    try {
      await _repo.delete(_groupId, id);
      state = state.copyWith(
        records: state.records.where((r) => r.id != id).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  String _fmt(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-'
      '${d.month.toString().padLeft(2, '0')}-'
      '${d.day.toString().padLeft(2, '0')}';
}

final attendanceDateProvider = StateNotifierProvider.family<
    AttendanceDateNotifier, AttendanceDateState, String>((ref, groupId) {
  return AttendanceDateNotifier(
    ref.watch(attendanceRepositoryProvider(groupId)),
    groupId,
  );
});

// ─── Student attendance history ───────────────────────────────────────────────

final studentAttendanceProvider = StateNotifierProvider.autoDispose.family<
    _StudentAttendanceNotifier,
    AsyncValue<List<Attendance>>,
    ({String groupId, String studentId})>((ref, params) {
  return _StudentAttendanceNotifier(
    ref.watch(attendanceRepositoryProvider(params.groupId)),
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
