import '../entities/attendance.dart';

abstract class AttendanceRepository {
  Future<List<RosterEntry>> roster(String groupId, String date);
  Future<List<Attendance>> listStudent(String groupId, String studentId);
  Future<Attendance> mark({
    required String groupId,
    required String studentId,
    required String lessonDate,
    required AttendanceStatus status,
  });
  Future<List<Attendance>> markMany({
    required String groupId,
    required String lessonDate,
    required Map<String, AttendanceStatus> marks,
  });
  Future<void> delete(String groupId, String id);
}
