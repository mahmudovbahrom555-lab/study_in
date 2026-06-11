import 'package:dio/dio.dart';

import '../domain/entities/attendance.dart';
import 'models/attendance_dto.dart';

class AttendanceApi {
  AttendanceApi(this._dio);

  final Dio _dio;

  Future<List<AttendanceDto>> listByDate(String groupId, String date) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/attendance',
      queryParameters: {'date': date},
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => AttendanceDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<AttendanceDto>> listStudent(
      String groupId, String studentId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/students/$studentId/attendance',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => AttendanceDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<AttendanceDto> mark({
    required String groupId,
    required String studentId,
    required String lessonDate,
    required AttendanceStatus status,
    String? note,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/attendance',
      data: {
        'student_id': studentId,
        'lesson_date': lessonDate,
        'status': status.toJson(),
        if (note != null) 'note': note,
      },
    );
    return AttendanceDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> delete(String groupId, String id) =>
      _dio.delete<void>('/groups/$groupId/attendance/$id');
}
