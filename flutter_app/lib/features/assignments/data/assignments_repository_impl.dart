import '../domain/entities/assignment.dart';
import '../domain/repositories/assignments_repository.dart';
import 'assignments_api.dart';

class AssignmentsRepositoryImpl implements AssignmentsRepository {
  AssignmentsRepositoryImpl(this._api);

  final AssignmentsApi _api;

  @override
  Future<List<Assignment>> listAssignments(String groupId) async {
    final dtos = await _api.listAssignments(groupId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<Assignment> createAssignment({
    required String groupId,
    required String title,
    String? description,
    DateTime? dueDate,
  }) async {
    final dto = await _api.createAssignment(
      groupId: groupId,
      title: title,
      description: description,
      dueDate: dueDate,
    );
    return dto.toDomain();
  }

  @override
  Future<void> submitAssignment(String groupId, String assignmentId,
          {String? comment}) =>
      _api.submitAssignment(groupId, assignmentId, comment: comment);

  @override
  Future<void> deleteAssignment(String groupId, String assignmentId) =>
      _api.deleteAssignment(groupId, assignmentId);
}
