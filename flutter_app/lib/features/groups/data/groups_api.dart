import 'package:dio/dio.dart';

import '../domain/entities/join_request.dart';
import 'models/group_dto.dart';
import 'models/join_request_dto.dart';

class GroupsApi {
  GroupsApi(this._dio);

  final Dio _dio;

  Future<List<GroupDto>> listMyGroups() async {
    final resp = await _dio.get<Map<String, dynamic>>('/groups');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GroupDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<GroupDto> getGroup(String id) async {
    final resp = await _dio.get<Map<String, dynamic>>('/groups/$id');
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<GroupDto> createGroup({
    required String name,
    String? subject,
    String? description,
    bool isPublic = false,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups',
      data: {
        'name': name,
        if (subject != null) 'subject': subject,
        if (description != null) 'description': description,
        'is_public': isPublic,
      },
    );
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<GroupDto> updateGroup(
    String id, {
    String? name,
    String? subject,
    String? description,
    bool? isPublic,
  }) async {
    final resp = await _dio.patch<Map<String, dynamic>>(
      '/groups/$id',
      data: {
        if (name != null) 'name': name,
        if (subject != null) 'subject': subject,
        if (description != null) 'description': description,
        if (isPublic != null) 'is_public': isPublic,
      },
    );
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> deleteGroup(String id) => _dio.delete<void>('/groups/$id');

  Future<GroupDto> archiveGroup(String id) async {
    final resp = await _dio.post<Map<String, dynamic>>('/groups/$id/archive');
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  /// Заявка в группу; в группу ученик попадёт, когда репетитор её примет.
  Future<JoinRequest> joinGroup(String inviteCode) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/join',
      data: {'invite_code': inviteCode},
    );
    return joinRequestFromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  /// Родитель записывает своего ребёнка: заявка сразу у репетитора.
  Future<JoinRequest> enrollChild(String inviteCode, String studentId) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/enroll',
      data: {'invite_code': inviteCode, 'student_id': studentId},
    );
    return joinRequestFromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<GroupPreview> previewGroup(String inviteCode) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/preview/$inviteCode',
    );
    return groupPreviewFromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  /// Открытые группы по названию, предмету или @username репетитора.
  Future<List<GroupPreview>> searchGroups(String query) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/search',
      queryParameters: {'q': query},
    );
    final list = resp.data!['data'] as List<dynamic>? ?? const [];
    return list
        .map((e) => groupPreviewFromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// Заявки детей, ждущие согласия родителя (вызывает родитель).
  Future<List<ParentConsent>> parentConsents() async {
    final resp = await _dio.get<Map<String, dynamic>>('/groups/consents');
    final list = resp.data!['data'] as List<dynamic>? ?? const [];
    return list
        .map((e) => parentConsentFromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> decideParentConsent(
    String requestId, {
    required bool approve,
  }) async {
    final action = approve ? 'approve' : 'decline';
    await _dio.post<void>('/groups/consents/$requestId/$action');
  }

  Future<List<JoinRequest>> myJoinRequests() async {
    final resp = await _dio.get<Map<String, dynamic>>('/groups/requests/mine');
    final list = resp.data!['data'] as List<dynamic>? ?? const [];
    return list
        .map((e) => joinRequestFromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<JoinRequestItem>> listJoinRequests(String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/requests',
    );
    final list = resp.data!['data'] as List<dynamic>? ?? const [];
    return list
        .map((e) => joinRequestItemFromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> decideJoinRequest(
    String groupId,
    String requestId, {
    required bool accept,
  }) async {
    final action = accept ? 'accept' : 'decline';
    await _dio.post<void>('/groups/$groupId/requests/$requestId/$action');
  }

  Future<void> leaveGroup(String groupId) =>
      _dio.post<void>('/groups/$groupId/members/leave');

  Future<List<GroupMemberDto>> listMembers(String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/members',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GroupMemberDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> removeMember(String groupId, String studentId) =>
      _dio.delete<void>('/groups/$groupId/members/$studentId');

  Future<void> updatePayment(String groupId, String studentId, String status) =>
      _dio.patch<void>(
        '/groups/$groupId/members/$studentId/payment',
        data: {'status': status},
      );
}
