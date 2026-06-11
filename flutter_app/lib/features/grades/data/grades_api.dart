import 'package:dio/dio.dart';

import 'models/grade_dto.dart';

class GradesApi {
  GradesApi(this._dio);

  final Dio _dio;

  Future<List<GradeDto>> listGroupGrades(String groupId) async {
    final resp =
        await _dio.get<Map<String, dynamic>>('/groups/$groupId/grades');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GradeDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<GradeDto>> listStudentGrades(
      String groupId, String studentId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/students/$studentId/grades',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => GradeDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<GradeDto> createGrade({
    required String groupId,
    required String studentId,
    required String subject,
    required double value,
    double maxValue = 100,
    String? comment,
    String? gradedAt,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/grades',
      data: {
        'student_id': studentId,
        'subject': subject,
        'value': value,
        'max_value': maxValue,
        if (comment != null) 'comment': comment,
        if (gradedAt != null) 'graded_at': gradedAt,
      },
    );
    return GradeDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<GradeDto> updateGrade(
    String gradeId, {
    String? subject,
    double? value,
    double? maxValue,
    String? comment,
    String? gradedAt,
  }) async {
    final resp = await _dio.patch<Map<String, dynamic>>(
      '/grades/$gradeId',
      data: {
        if (subject != null) 'subject': subject,
        if (value != null) 'value': value,
        if (maxValue != null) 'max_value': maxValue,
        if (comment != null) 'comment': comment,
        if (gradedAt != null) 'graded_at': gradedAt,
      },
    );
    return GradeDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> deleteGrade(String gradeId) =>
      _dio.delete<void>('/grades/$gradeId');
}
