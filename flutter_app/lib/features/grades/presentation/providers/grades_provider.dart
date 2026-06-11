import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/grades_api.dart';
import '../../data/grades_repository_impl.dart';
import '../../domain/entities/grade.dart';
import '../../domain/repositories/grades_repository.dart';

final gradesRepositoryProvider =
    Provider.family<GradesRepository, String>((ref, groupId) {
  final dio = ref.watch(dioProvider);
  return GradesRepositoryImpl(GradesApi(dio));
});

// ─── Grades list state ────────────────────────────────────────────────────────

class GradesState {
  const GradesState({
    this.grades = const [],
    this.isLoading = false,
    this.error,
  });

  final List<Grade> grades;
  final bool isLoading;
  final String? error;

  GradesState copyWith({
    List<Grade>? grades,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      GradesState(
        grades: grades ?? this.grades,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

// ─── Group-level grades (teacher view) ───────────────────────────────────────

class GroupGradesNotifier extends StateNotifier<GradesState> {
  GroupGradesNotifier(this._repo, this._groupId) : super(const GradesState());

  final GradesRepository _repo;
  final String _groupId;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final gs = await _repo.listGroupGrades(_groupId);
      state = state.copyWith(isLoading: false, grades: gs);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> createGrade({
    required String studentId,
    required String subject,
    required double value,
    double maxValue = 100,
    String? comment,
  }) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final g = await _repo.createGrade(
        groupId: _groupId,
        studentId: studentId,
        subject: subject,
        value: value,
        maxValue: maxValue,
        comment: comment,
      );
      state = state.copyWith(
        isLoading: false,
        grades: [g, ...state.grades],
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> deleteGrade(String gradeId) async {
    try {
      await _repo.deleteGrade(gradeId);
      state = state.copyWith(
        grades: state.grades.where((g) => g.id != gradeId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

final groupGradesProvider = StateNotifierProvider.family<GroupGradesNotifier,
    GradesState, String>((ref, groupId) {
  return GroupGradesNotifier(
    ref.watch(gradesRepositoryProvider(groupId)),
    groupId,
  );
});

// ─── Student grades (student self-view) ───────────────────────────────────────

class StudentGradesNotifier extends StateNotifier<AsyncValue<List<Grade>>> {
  StudentGradesNotifier(this._repo, this._groupId, this._studentId)
      : super(const AsyncValue.loading()) {
    load();
  }

  final GradesRepository _repo;
  final String _groupId;
  final String _studentId;

  Future<void> load() async {
    state = const AsyncValue.loading();
    try {
      final gs = await _repo.listStudentGrades(_groupId, _studentId);
      state = AsyncValue.data(gs);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}

final studentGradesProvider = StateNotifierProvider.family<StudentGradesNotifier,
    AsyncValue<List<Grade>>, ({String groupId, String studentId})>(
  (ref, params) => StudentGradesNotifier(
    ref.watch(gradesRepositoryProvider(params.groupId)),
    params.groupId,
    params.studentId,
  ),
);
