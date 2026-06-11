import '../../../attendance/domain/entities/attendance.dart';
import '../../../grades/domain/entities/grade.dart';
import '../../../groups/domain/entities/group.dart';
import '../domain/entities/parent_link.dart';
import '../domain/repositories/parents_repository.dart';
import 'parents_api.dart';

class ParentsRepositoryImpl implements ParentsRepository {
  ParentsRepositoryImpl(this._api);

  final ParentsApi _api;

  @override
  Future<ParentLink> linkChild(String studentId) async {
    final dto = await _api.linkChild(studentId);
    return dto.toDomain();
  }

  @override
  Future<void> unlinkChild(String studentId) => _api.unlinkChild(studentId);

  @override
  Future<List<ParentLink>> listChildren() async {
    final dtos = await _api.listChildren();
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<List<Group>> childGroups(String studentId) async {
    final dtos = await _api.childGroups(studentId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<List<Grade>> childGrades(String studentId, String groupId) async {
    final dtos = await _api.childGrades(studentId, groupId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<List<Attendance>> childAttendance(
      String studentId, String groupId) async {
    final dtos = await _api.childAttendance(studentId, groupId);
    return dtos.map((d) => d.toDomain()).toList();
  }
}
