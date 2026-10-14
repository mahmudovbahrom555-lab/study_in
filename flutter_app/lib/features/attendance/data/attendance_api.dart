import 'package:dio/dio.dart';

import '../domain/entities/attendance.dart';
import 'models/attendance_dto.dart';

class AttendanceApi {
  AttendanceApi(this._dio);

  final Dio _dio;

  /// Все ученики группы на дату и их отметки.
  Future<List<RosterEntry>> roster(String groupId, String date) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/attendance',
      queryParameters: {'date': date},
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => AttendanceDto.rosterFromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<AttendanceDto>> listStudent(
    String groupId,
    String studentId,
  ) async {
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
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/attendance',
      data: {
        'student_id': studentId,
        'lesson_date': lessonDate,
        'status': status.toJson(),
      },
    );
    return AttendanceDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  /// Несколько отметок за день одним запросом — сохраняются все или ни одной.
  Future<List<AttendanceDto>> markMany({
    required String groupId,
    required String lessonDate,
    required Map<String, AttendanceStatus> marks,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/attendance/batch',
      data: {
        'lesson_date': lessonDate,
        'marks': [
          for (final m in marks.entries)
            {'student_id': m.key, 'status': m.value.toJson()},
        ],
      },
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => AttendanceDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> delete(String groupId, String id) =>
      _dio.delete<void>('/groups/$groupId/attendance/$id');
}
