import 'package:dio/dio.dart';

import 'models/group_dto.dart';

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
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups',
      data: {
        'name': name,
        if (subject != null) 'subject': subject,
        if (description != null) 'description': description,
      },
    );
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<GroupDto> updateGroup(
    String id, {
    String? name,
    String? subject,
    String? description,
  }) async {
    final resp = await _dio.patch<Map<String, dynamic>>(
      '/groups/$id',
      data: {
        if (name != null) 'name': name,
        if (subject != null) 'subject': subject,
        if (description != null) 'description': description,
      },
    );
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> deleteGroup(String id) =>
      _dio.delete<void>('/groups/$id');

  Future<GroupDto> archiveGroup(String id) async {
    final resp =
        await _dio.post<Map<String, dynamic>>('/groups/$id/archive');
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<GroupDto> joinGroup(String inviteCode) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/join',
      data: {'invite_code': inviteCode},
    );
    return GroupDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> leaveGroup(String groupId) =>
      _dio.post<void>('/groups/$groupId/members/leave');

  Future<List<GroupMemberDto>> listMembers(String groupId) async {
    final resp =
        await _dio.get<Map<String, dynamic>>('/groups/$groupId/members');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GroupMemberDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> removeMember(String groupId, String studentId) =>
      _dio.delete<void>('/groups/$groupId/members/$studentId');

  Future<void> updatePayment(
    String groupId,
    String studentId,
    String status,
  ) =>
      _dio.patch<void>(
        '/groups/$groupId/members/$studentId/payment',
        data: {'status': status},
      );
}
