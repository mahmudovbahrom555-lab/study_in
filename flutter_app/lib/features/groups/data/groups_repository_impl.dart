import '../domain/entities/group.dart';
import '../domain/entities/join_request.dart';
import '../domain/repositories/groups_repository.dart';
import 'groups_api.dart';

class GroupsRepositoryImpl implements GroupsRepository {
  GroupsRepositoryImpl(this._api);

  final GroupsApi _api;

  @override
  Future<List<Group>> listMyGroups() async {
    final dtos = await _api.listMyGroups();
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<Group> getGroup(String id) async {
    final dto = await _api.getGroup(id);
    return dto.toDomain();
  }

  @override
  Future<Group> createGroup({
    required String name,
    String? subject,
    String? description,
    bool isPublic = false,
  }) async {
    final dto = await _api.createGroup(
      name: name,
      subject: subject,
      description: description,
      isPublic: isPublic,
    );
    return dto.toDomain();
  }

  @override
  Future<Group> updateGroup(
    String id, {
    String? name,
    String? subject,
    String? description,
    bool? isPublic,
  }) async {
    final dto = await _api.updateGroup(
      id,
      name: name,
      subject: subject,
      description: description,
      isPublic: isPublic,
    );
    return dto.toDomain();
  }

  @override
  Future<void> deleteGroup(String id) => _api.deleteGroup(id);

  @override
  Future<Group> archiveGroup(String id) async {
    final dto = await _api.archiveGroup(id);
    return dto.toDomain();
  }

  @override
  Future<JoinRequest> joinGroup(String inviteCode) =>
      _api.joinGroup(inviteCode);

  @override
  Future<GroupPreview> previewGroup(String inviteCode) =>
      _api.previewGroup(inviteCode);

  @override
  Future<List<GroupPreview>> searchGroups(String query) =>
      _api.searchGroups(query);

  @override
  Future<List<ParentConsent>> parentConsents() => _api.parentConsents();

  @override
  Future<void> decideParentConsent(
    String requestId, {
    required bool approve,
  }) => _api.decideParentConsent(requestId, approve: approve);

  @override
  Future<List<JoinRequest>> myJoinRequests() => _api.myJoinRequests();

  @override
  Future<List<JoinRequestItem>> listJoinRequests(String groupId) =>
      _api.listJoinRequests(groupId);

  @override
  Future<void> decideJoinRequest(
    String groupId,
    String requestId, {
    required bool accept,
  }) => _api.decideJoinRequest(groupId, requestId, accept: accept);

  @override
  Future<void> leaveGroup(String groupId) => _api.leaveGroup(groupId);

  @override
  Future<List<GroupMember>> listMembers(String groupId) async {
    final dtos = await _api.listMembers(groupId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<void> removeMember(String groupId, String studentId) =>
      _api.removeMember(groupId, studentId);

  @override
  Future<void> updatePayment(
    String groupId,
    String studentId,
    PaymentStatus status,
  ) => _api.updatePayment(groupId, studentId, status.toJson());
}
