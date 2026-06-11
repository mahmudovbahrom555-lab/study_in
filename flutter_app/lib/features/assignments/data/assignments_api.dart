import 'package:dio/dio.dart';

import 'models/assignment_dto.dart';

class AssignmentsApi {
  AssignmentsApi(this._dio);

  final Dio _dio;

  Future<List<AssignmentDto>> listAssignments(String groupId) async {
    final resp = await _dio
        .get<Map<String, dynamic>>('/groups/$groupId/assignments');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => AssignmentDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<AssignmentDto> createAssignment({
    required String groupId,
    required String title,
    String? description,
    DateTime? dueDate,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/assignments',
      data: {
        'title': title,
        if (description != null) 'description': description,
        if (dueDate != null) 'due_date': dueDate.toIso8601String(),
      },
    );
    return AssignmentDto.fromJson(
        resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> submitAssignment(String groupId, String assignmentId,
      {String? comment}) async {
    await _dio.post<void>(
      '/groups/$groupId/assignments/$assignmentId/submit',
      data: {if (comment != null) 'comment': comment},
    );
  }

  Future<void> deleteAssignment(String groupId, String assignmentId) =>
      _dio.delete<void>('/groups/$groupId/assignments/$assignmentId');
}
