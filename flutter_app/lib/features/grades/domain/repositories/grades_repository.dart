import '../entities/grade.dart';

abstract class GradesRepository {
  Future<List<Grade>> listGroupGrades(String groupId);
  Future<List<Grade>> listStudentGrades(String groupId, String studentId);
  Future<Grade> createGrade({
    required String groupId,
    required String studentId,
    required String subject,
    required double value,
    double maxValue,
    String? comment,
  });
  Future<void> deleteGrade(String groupId, String gradeId);
}
