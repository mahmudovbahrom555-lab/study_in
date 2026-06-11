import '../domain/entities/grade.dart';
import '../domain/repositories/grades_repository.dart';
import 'grades_api.dart';

class GradesRepositoryImpl implements GradesRepository {
  GradesRepositoryImpl(this._api);

  final GradesApi _api;

  @override
  Future<List<Grade>> listGroupGrades(String groupId) async {
    final dtos = await _api.listGroupGrades(groupId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<List<Grade>> listStudentGrades(String groupId, String studentId) async {
    final dtos = await _api.listStudentGrades(groupId, studentId);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<Grade> createGrade({
    required String groupId,
    required String studentId,
    required String subject,
    required double value,
    double maxValue = 100,
    String? comment,
    String? gradedAt,
  }) async {
    final dto = await _api.createGrade(
      groupId: groupId,
      studentId: studentId,
      subject: subject,
      value: value,
      maxValue: maxValue,
      comment: comment,
      gradedAt: gradedAt,
    );
    return dto.toDomain();
  }

  @override
  Future<Grade> updateGrade(
    String gradeId, {
    String? subject,
    double? value,
    double? maxValue,
    String? comment,
    String? gradedAt,
  }) async {
    final dto = await _api.updateGrade(
      gradeId,
      subject: subject,
      value: value,
      maxValue: maxValue,
      comment: comment,
      gradedAt: gradedAt,
    );
    return dto.toDomain();
  }

  @override
  Future<void> deleteGrade(String gradeId) => _api.deleteGrade(gradeId);
}
