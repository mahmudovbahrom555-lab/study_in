import 'package:dio/dio.dart';

import '../../../attendance/data/models/attendance_dto.dart';
import '../../../grades/data/models/grade_dto.dart';
import '../../../groups/data/models/group_dto.dart';
import 'models/parent_link_dto.dart';

class ParentsApi {
  ParentsApi(this._dio);

  final Dio _dio;

  Future<ParentLinkDto> linkChild(String studentId) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/parent/children',
      data: {'student_id': studentId},
    );
    return ParentLinkDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> unlinkChild(String studentId) =>
      _dio.delete<void>('/parent/children/$studentId');

  Future<List<ParentLinkDto>> listChildren() async {
    final resp =
        await _dio.get<Map<String, dynamic>>('/parent/children');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => ParentLinkDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<GroupDto>> childGroups(String studentId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/parent/children/$studentId/groups',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GroupDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<GradeDto>> childGrades(String studentId, String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/parent/children/$studentId/groups/$groupId/grades',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GradeDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<AttendanceDto>> childAttendance(
      String studentId, String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/parent/children/$studentId/groups/$groupId/attendance',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => AttendanceDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
