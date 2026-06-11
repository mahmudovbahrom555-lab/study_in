import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../../attendance/domain/entities/attendance.dart';
import '../../../grades/domain/entities/grade.dart';
import '../../../groups/domain/entities/group.dart';
import '../../data/parents_api.dart';
import '../../data/parents_repository_impl.dart';
import '../../domain/entities/parent_link.dart';
import '../../domain/repositories/parents_repository.dart';

final parentsRepositoryProvider = Provider<ParentsRepository>((ref) {
  final dio = ref.watch(dioProvider);
  return ParentsRepositoryImpl(ParentsApi(dio));
});

// ─── Children list ────────────────────────────────────────────────────────────

class ChildrenState {
  const ChildrenState({
    this.children = const [],
    this.isLoading = false,
    this.error,
  });

  final List<ParentLink> children;
  final bool isLoading;
  final String? error;

  ChildrenState copyWith({
    List<ParentLink>? children,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      ChildrenState(
        children: children ?? this.children,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

class ChildrenNotifier extends StateNotifier<ChildrenState> {
  ChildrenNotifier(this._repo) : super(const ChildrenState());

  final ParentsRepository _repo;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final children = await _repo.listChildren();
      state = state.copyWith(isLoading: false, children: children);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> linkChild(String studentId) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final link = await _repo.linkChild(studentId);
      state = state.copyWith(
        isLoading: false,
        children: [link, ...state.children],
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> unlinkChild(String studentId) async {
    try {
      await _repo.unlinkChild(studentId);
      state = state.copyWith(
        children: state.children
            .where((c) => c.studentId != studentId)
            .toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

final childrenProvider =
    StateNotifierProvider<ChildrenNotifier, ChildrenState>((ref) {
  return ChildrenNotifier(ref.watch(parentsRepositoryProvider));
});

// ─── Child groups ─────────────────────────────────────────────────────────────

final childGroupsProvider =
    FutureProvider.family<List<Group>, String>((ref, studentId) async {
  return ref.watch(parentsRepositoryProvider).childGroups(studentId);
});

// ─── Child grades per group ───────────────────────────────────────────────────

final childGradesProvider = FutureProvider.family<List<Grade>,
    ({String studentId, String groupId})>((ref, params) async {
  return ref
      .watch(parentsRepositoryProvider)
      .childGrades(params.studentId, params.groupId);
});

// ─── Child attendance per group ───────────────────────────────────────────────

final childAttendanceProvider = FutureProvider.family<List<Attendance>,
    ({String studentId, String groupId})>((ref, params) async {
  return ref
      .watch(parentsRepositoryProvider)
      .childAttendance(params.studentId, params.groupId);
});
