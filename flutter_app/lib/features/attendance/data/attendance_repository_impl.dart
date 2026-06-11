import '../domain/entities/attendance.dart';
import '../domain/repositories/attendance_repository.dart';
import 'attendance_api.dart';

class AttendanceRepositoryImpl implements AttendanceRepository {
  AttendanceRepositoryImpl(this._api);

  final AttendanceApi _api;

  @override
  Future<List<Attendance>> listByDate(String groupId, String date) async {
    final dtos = await _api.listByDate(groupId, date);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<List<Attendance>> listStudent(String groupId, String studentId) async {
    final dtos = await _api.listStudent(groupId, studentId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<Attendance> mark({
    required String groupId,
    required String studentId,
    required String lessonDate,
    required AttendanceStatus status,
    String? note,
  }) async {
    final dto = await _api.mark(
      groupId: groupId,
      studentId: studentId,
      lessonDate: lessonDate,
      status: status,
      note: note,
    );
    return dto.toDomain();
  }

  @override
  Future<void> delete(String groupId, String id) =>
      _api.delete(groupId, id);
}
