import '../entities/attendance.dart';

abstract class AttendanceRepository {
  Future<List<Attendance>> listByDate(String groupId, String date);
  Future<List<Attendance>> listStudent(String groupId, String studentId);
  Future<Attendance> mark({
    required String groupId,
    required String studentId,
    required String lessonDate,
    required AttendanceStatus status,
    String? note,
  });
  Future<void> delete(String groupId, String id);
}
