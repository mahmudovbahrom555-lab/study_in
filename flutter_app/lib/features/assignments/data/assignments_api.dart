import 'package:dio/dio.dart';

import '../domain/entities/assignment.dart';
import '../domain/entities/submission.dart';
import 'models/assignment_dto.dart';

class AssignmentsApi {
  AssignmentsApi(this._dio);

  final Dio _dio;

  String _base(String groupId) => '/groups/$groupId/assignments';

  Future<List<Assignment>> listAssignments(String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>(_base(groupId));
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => assignmentFromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Assignment> createAssignment({
    required String groupId,
    required String title,
    String? description,
    DateTime? dueDate,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      _base(groupId),
      data: {
        'title': title,
        if (description != null) 'description': description,
        if (dueDate != null) 'due_date': dueDate.toUtc().toIso8601String(),
      },
    );
    return assignmentFromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> deleteAssignment(String groupId, String assignmentId) =>
      _dio.delete<void>('${_base(groupId)}/$assignmentId');

  /// Загружает файл для будущей сдачи. До сдачи он «ожидает» и через сутки
  /// удаляется сервером, если так и не попал в сдачу.
  Future<AttachedFile> uploadFile({
    required String path,
    required String name,
    void Function(int sent, int total)? onProgress,
    CancelToken? cancelToken,
  }) async {
    final form = FormData.fromMap({
      'purpose': 'submission',
      'file': await MultipartFile.fromFile(path, filename: name),
    });
    final resp = await _dio.post<Map<String, dynamic>>(
      '/files',
      data: form,
      onSendProgress: onProgress,
      cancelToken: cancelToken,
    );
    return attachedFileFromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  /// Сдаёт или пересдаёт работу; [fileIds] — полный список, он заменяет прежний.
  Future<Submission> submit(
    String groupId,
    String assignmentId, {
    String? comment,
    required List<String> fileIds,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '${_base(groupId)}/$assignmentId/submit',
      data: {
        if (comment != null) 'comment': comment,
        'file_ids': fileIds,
      },
    );
    return submissionFromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  /// Своя сдача ученика; null — ещё не сдавал.
  Future<Submission?> mySubmission(String groupId, String assignmentId) async {
    try {
      final resp = await _dio.get<Map<String, dynamic>>(
        '${_base(groupId)}/$assignmentId/submission',
      );
      return submissionFromJson(resp.data!['data'] as Map<String, dynamic>);
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<List<Submission>> listSubmissions(
    String groupId,
    String assignmentId,
  ) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '${_base(groupId)}/$assignmentId/submissions',
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => submissionFromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> grade(
    String groupId,
    String assignmentId,
    String submissionId, {
    required int grade,
    String? note,
  }) =>
      _dio.post<void>(
        '${_base(groupId)}/$assignmentId/submissions/$submissionId/grade',
        data: {'grade': grade, if (note != null) 'note': note},
      );
}
