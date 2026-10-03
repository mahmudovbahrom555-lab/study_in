import '../entities/parent_link.dart';
import '../../../grades/domain/entities/grade.dart';
import '../../../attendance/domain/entities/attendance.dart';
import '../../../groups/domain/entities/group.dart';

abstract class ParentsRepository {
  /// Привязка по коду, который ученик получил в своём аккаунте.
  Future<ParentLink> linkChildByCode(String code);

  /// Ученик: выдать код для привязки родителя (действует 24 часа).
  Future<({String code, DateTime expiresAt})> createLinkCode();
  Future<void> unlinkChild(String studentId);
  Future<List<ParentLink>> listChildren();
  Future<List<Group>> childGroups(String studentId);
  Future<List<Grade>> childGrades(String studentId, String groupId);
  Future<List<Attendance>> childAttendance(String studentId, String groupId);
}
