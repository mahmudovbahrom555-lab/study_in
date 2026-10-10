import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/groups_api.dart';
import '../../data/groups_repository_impl.dart';
import '../../domain/entities/group.dart';
import '../../domain/repositories/groups_repository.dart';
import '../../../../core/network/error_message.dart';

// ─── Providers ───────────────────────────────────────────────────────────────

final groupsRepositoryProvider = Provider<GroupsRepository>((ref) {
  final dio = ref.watch(dioProvider);
  return GroupsRepositoryImpl(GroupsApi(dio));
});

// ─── Groups list ─────────────────────────────────────────────────────────────

class GroupsState {
  const GroupsState({
    this.groups = const [],
    this.isLoading = false,
    this.error,
  });

  final List<Group> groups;
  final bool isLoading;
  final String? error;

  GroupsState copyWith({
    List<Group>? groups,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) =>
      GroupsState(
        groups: groups ?? this.groups,
        isLoading: isLoading ?? this.isLoading,
        error: clearError ? null : (error ?? this.error),
      );
}

class GroupsNotifier extends StateNotifier<GroupsState> {
  GroupsNotifier(this._repo) : super(const GroupsState());

  final GroupsRepository _repo;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final gs = await _repo.listMyGroups();
      state = state.copyWith(isLoading: false, groups: gs);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: userErrorMessage(e));
    }
  }

  Future<void> createGroup({
    required String name,
    String? subject,
    String? description,
  }) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final g = await _repo.createGroup(
        name: name,
        subject: subject,
        description: description,
      );
      state = state.copyWith(
        isLoading: false,
        groups: [g, ...state.groups],
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: userErrorMessage(e));
    }
  }

  Future<void> joinGroup(String inviteCode) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final g = await _repo.joinGroup(inviteCode);
      state = state.copyWith(
        isLoading: false,
        groups: [g, ...state.groups],
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: userErrorMessage(e));
    }
  }

  Future<void> updateGroup(
    String id, {
    required String name,
    required String subject,
    required String description,
  }) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final updated = await _repo.updateGroup(
        id,
        name: name,
        subject: subject,
        description: description,
      );
      state = state.copyWith(
        isLoading: false,
        groups: state.groups.map((g) => g.id == id ? updated : g).toList(),
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: userErrorMessage(e));
    }
  }

  Future<void> archiveGroup(String id) async {
    try {
      final updated = await _repo.archiveGroup(id);
      state = state.copyWith(
        groups: state.groups.map((g) => g.id == id ? updated : g).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: userErrorMessage(e));
    }
  }

  Future<void> deleteGroup(String id) async {
    try {
      await _repo.deleteGroup(id);
      state = state.copyWith(
        groups: state.groups.where((g) => g.id != id).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: userErrorMessage(e));
    }
  }
}

final groupsProvider =
    StateNotifierProvider<GroupsNotifier, GroupsState>((ref) {
  return GroupsNotifier(ref.watch(groupsRepositoryProvider));
});

// ─── Members list ─────────────────────────────────────────────────────────────

class MembersNotifier extends StateNotifier<AsyncValue<List<GroupMember>>> {
  MembersNotifier(this._repo, this._groupId)
      : super(const AsyncValue.loading()) {
    load();
  }

  final GroupsRepository _repo;
  final String _groupId;

  Future<void> load() async {
    state = const AsyncValue.loading();
    try {
      final members = await _repo.listMembers(_groupId);
      state = AsyncValue.data(members);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> removeMember(String studentId) async {
    await _repo.removeMember(_groupId, studentId);
    final current = state.valueOrNull ?? [];
    state = AsyncValue.data(
      current.where((m) => m.studentId != studentId).toList(),
    );
  }

  Future<void> updatePayment(String studentId, PaymentStatus status) async {
    await _repo.updatePayment(_groupId, studentId, status);
    final current = state.valueOrNull ?? [];
    state = AsyncValue.data(
      current
          .map(
            (m) => m.studentId == studentId
                ? GroupMember(
                    studentId: m.studentId,
                    name: m.name,
                    phone: m.phone,
                    avatarUrl: m.avatarUrl,
                    joinedAt: m.joinedAt,
                    paymentStatus: status,
                  )
                : m,
          )
          .toList(),
    );
  }
}

final membersProvider = StateNotifierProvider.family<MembersNotifier,
    AsyncValue<List<GroupMember>>, String>((ref, groupId) {
  return MembersNotifier(ref.watch(groupsRepositoryProvider), groupId);
});
