import '../entities/group.dart';

abstract class GroupsRepository {
  Future<List<Group>> listMyGroups();
  Future<Group> getGroup(String id);
  Future<Group> createGroup({
    required String name,
    String? subject,
    String? description,
  });
  Future<Group> updateGroup(
    String id, {
    String? name,
    String? subject,
    String? description,
  });
  Future<void> deleteGroup(String id);
  Future<Group> archiveGroup(String id);
  Future<Group> joinGroup(String inviteCode);
  Future<void> leaveGroup(String groupId);
  Future<List<GroupMember>> listMembers(String groupId);
  Future<void> removeMember(String groupId, String studentId);
  Future<void> updatePayment(
    String groupId,
    String studentId,
    PaymentStatus status,
  );
}
