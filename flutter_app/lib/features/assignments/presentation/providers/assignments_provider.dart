import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/assignments_api.dart';
import '../../data/assignments_repository_impl.dart';
import '../../domain/entities/assignment.dart';
import '../../domain/repositories/assignments_repository.dart';

final assignmentsRepositoryProvider =
    Provider.family<AssignmentsRepository, String>((ref, groupId) {
  final dio = ref.watch(dioProvider);
  return AssignmentsRepositoryImpl(AssignmentsApi(dio));
});

// ─── State ────────────────────────────────────────────────────────────────────

class AssignmentsState {
  const AssignmentsState({
    this.assignments = const [],
    this.isLoading = false,
    this.error,
  });

  final List<Assignment> assignments;
  final bool isLoading;
  final String? error;

  AssignmentsState copyWith({
    List<Assignment>? assignments,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      AssignmentsState(
        assignments: assignments ?? this.assignments,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

// ─── Notifier ─────────────────────────────────────────────────────────────────

class AssignmentsNotifier extends StateNotifier<AssignmentsState> {
  AssignmentsNotifier(this._repo, this._groupId)
      : super(const AssignmentsState());

  final AssignmentsRepository _repo;
  final String _groupId;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final items = await _repo.listAssignments(_groupId);
      state = state.copyWith(isLoading: false, assignments: items);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> create({
    required String title,
    String? description,
    DateTime? dueDate,
  }) async {
    try {
      final item = await _repo.createAssignment(
        groupId: _groupId,
        title: title,
        description: description,
        dueDate: dueDate,
      );
      state = state.copyWith(
        assignments: [item, ...state.assignments],
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> submit(String assignmentId, {String? comment}) async {
    try {
      await _repo.submitAssignment(_groupId, assignmentId, comment: comment);
      await load();
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> delete(String assignmentId) async {
    try {
      await _repo.deleteAssignment(_groupId, assignmentId);
      state = state.copyWith(
        assignments: state.assignments
            .where((a) => a.id != assignmentId)
            .toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

final assignmentsProvider = StateNotifierProvider.family<AssignmentsNotifier,
    AssignmentsState, String>((ref, groupId) {
  return AssignmentsNotifier(
    ref.watch(assignmentsRepositoryProvider(groupId)),
    groupId,
  );
});
