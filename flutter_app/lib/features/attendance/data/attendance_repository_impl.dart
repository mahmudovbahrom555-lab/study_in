import '../domain/entities/attendance.dart';
import '../domain/repositories/attendance_repository.dart';
import 'attendance_api.dart';

class AttendanceRepositoryImpl implements AttendanceRepository {
  AttendanceRepositoryImpl(this._api);

  final AttendanceApi _api;

  @override
  Future<List<RosterEntry>> roster(String groupId, String date) =>
      _api.roster(groupId, date);

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
  }) async {
    final dto = await _api.mark(
      groupId: groupId,
      studentId: studentId,
      lessonDate: lessonDate,
      status: status,
    );
    return dto.toDomain();
  }

  @override
  Future<List<Attendance>> markMany({
    required String groupId,
    required String lessonDate,
    required Map<String, AttendanceStatus> marks,
  }) async {
    final dtos = await _api.markMany(
      groupId: groupId,
      lessonDate: lessonDate,
      marks: marks,
    );
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<void> delete(String groupId, String id) => _api.delete(groupId, id);
}
