import '../entities/report.dart';

abstract class ReportsRepository {
  Future<ParentReport> getParentRoi(String studentId, String groupId);
  Future<List<RiskAlert>> getRiskAlerts({String status = ''});
  Future<List<RiskAlert>> refreshRiskAlerts();
  Future<void> resolveAlert(String alertId);
}
