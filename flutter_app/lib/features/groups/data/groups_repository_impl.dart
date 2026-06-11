import '../domain/entities/group.dart';
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
  }) async {
    final dto = await _api.createGroup(
      name: name,
      subject: subject,
      description: description,
    );
    return dto.toDomain();
  }

  @override
  Future<Group> updateGroup(
    String id, {
    String? name,
    String? subject,
    String? description,
  }) async {
    final dto = await _api.updateGroup(
      id,
      name: name,
      subject: subject,
      description: description,
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
  Future<Group> joinGroup(String inviteCode) async {
    final dto = await _api.joinGroup(inviteCode);
    return dto.toDomain();
  }

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
  ) =>
      _api.updatePayment(groupId, studentId, status.toJson());
}
