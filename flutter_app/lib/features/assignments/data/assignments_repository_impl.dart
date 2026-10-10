import 'package:dio/dio.dart';

import '../domain/entities/assignment.dart';
import '../domain/entities/submission.dart';
import '../domain/repositories/assignments_repository.dart';
import 'assignments_api.dart';

class AssignmentsRepositoryImpl implements AssignmentsRepository {
  AssignmentsRepositoryImpl(this._api);

  final AssignmentsApi _api;

  @override
  Future<List<Assignment>> listAssignments(String groupId) =>
      _api.listAssignments(groupId);

  @override
  Future<Assignment> createAssignment({
    required String groupId,
    required String title,
    String? description,
    DateTime? dueDate,
  }) =>
      _api.createAssignment(
        groupId: groupId,
        title: title,
        description: description,
        dueDate: dueDate,
      );

  @override
  Future<void> deleteAssignment(String groupId, String assignmentId) =>
      _api.deleteAssignment(groupId, assignmentId);

  @override
  Future<AttachedFile> uploadFile({
    required String path,
    required String name,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancelToken,
  }) =>
      _api.uploadFile(
        path: path,
        name: name,
        onProgress: onProgress,
        cancelToken: cancelToken,
      );

  @override
  Future<Submission> submit(
    String groupId,
    String assignmentId, {
    String? comment,
    required List<String> fileIds,
  }) =>
      _api.submit(groupId, assignmentId, comment: comment, fileIds: fileIds);

  @override
  Future<Submission?> mySubmission(String groupId, String assignmentId) =>
      _api.mySubmission(groupId, assignmentId);

  @override
  Future<List<Submission>> listSubmissions(
    String groupId,
    String assignmentId,
  ) =>
      _api.listSubmissions(groupId, assignmentId);

  @override
  Future<void> grade(
    String groupId,
    String assignmentId,
    String submissionId, {
    required int grade,
    String? note,
  }) =>
      _api.grade(groupId, assignmentId, submissionId, grade: grade, note: note);
}
