import '../entities/group.dart';
import '../entities/join_request.dart';

abstract class GroupsRepository {
  Future<List<Group>> listMyGroups();
  Future<Group> getGroup(String id);
  Future<Group> createGroup({
    required String name,
    String? subject,
    String? description,
    bool isPublic = false,
  });
  Future<Group> updateGroup(
    String id, {
    String? name,
    String? subject,
    String? description,
    bool? isPublic,
  });
  Future<void> deleteGroup(String id);
  Future<Group> archiveGroup(String id);
  Future<JoinRequest> joinGroup(String inviteCode);
  Future<GroupPreview> previewGroup(String inviteCode);
  Future<List<GroupPreview>> searchGroups(String query);
  Future<List<JoinRequest>> myJoinRequests();
  Future<List<JoinRequestItem>> listJoinRequests(String groupId);
  Future<void> decideJoinRequest(
    String groupId,
    String requestId, {
    required bool accept,
  });
  Future<void> leaveGroup(String groupId);
  Future<List<GroupMember>> listMembers(String groupId);
  Future<void> removeMember(String groupId, String studentId);
  Future<void> updatePayment(
    String groupId,
    String studentId,
    PaymentStatus status,
  );
}
