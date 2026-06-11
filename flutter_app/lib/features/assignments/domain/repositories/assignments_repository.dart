import '../entities/assignment.dart';

abstract class AssignmentsRepository {
  Future<List<Assignment>> listAssignments(String groupId);
  Future<Assignment> createAssignment({
    required String groupId,
    required String title,
    String? description,
    DateTime? dueDate,
  });
  Future<void> submitAssignment(String groupId, String assignmentId,
      {String? comment});
  Future<void> deleteAssignment(String groupId, String assignmentId);
}
