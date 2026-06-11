import '../entities/parent_link.dart';
import '../../../grades/domain/entities/grade.dart';
import '../../../attendance/domain/entities/attendance.dart';
import '../../../groups/domain/entities/group.dart';

abstract class ParentsRepository {
  Future<ParentLink> linkChild(String studentId);
  Future<void> unlinkChild(String studentId);
  Future<List<ParentLink>> listChildren();
  Future<List<Group>> childGroups(String studentId);
  Future<List<Grade>> childGrades(String studentId, String groupId);
  Future<List<Attendance>> childAttendance(String studentId, String groupId);
}
