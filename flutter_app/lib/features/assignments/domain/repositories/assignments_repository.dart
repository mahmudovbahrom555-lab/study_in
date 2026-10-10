import 'package:dio/dio.dart';

import '../entities/assignment.dart';
import '../entities/submission.dart';

abstract class AssignmentsRepository {
  Future<List<Assignment>> listAssignments(String groupId);
  Future<Assignment> createAssignment({
    required String groupId,
    required String title,
    String? description,
    DateTime? dueDate,
  });
  Future<void> deleteAssignment(String groupId, String assignmentId);

  Future<AttachedFile> uploadFile({
    required String path,
    required String name,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancelToken,
  });
  Future<Submission> submit(
    String groupId,
    String assignmentId, {
    String? comment,
    required List<String> fileIds,
  });
  Future<Submission?> mySubmission(String groupId, String assignmentId);
  Future<List<Submission>> listSubmissions(String groupId, String assignmentId);
  Future<void> grade(
    String groupId,
    String assignmentId,
    String submissionId, {
    required int grade,
    String? note,
  });
}
