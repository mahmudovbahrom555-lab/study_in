import '../domain/entities/report.dart';
import '../domain/repositories/reports_repository.dart';
import 'reports_api.dart';

class ReportsRepositoryImpl implements ReportsRepository {
  ReportsRepositoryImpl(this._api);

  final ReportsApi _api;

  @override
  Future<ParentReport> getParentRoi(String studentId, String groupId) async {
    final dto = await _api.getParentRoi(studentId, groupId);
    return dto.toDomain();
  }

  @override
  Future<List<RiskAlert>> getRiskAlerts({String status = ''}) async {
    final dtos = await _api.getRiskAlerts(status: status);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<List<RiskAlert>> refreshRiskAlerts() async {
    final dtos = await _api.refreshRiskAlerts();
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<void> resolveAlert(String alertId) => _api.resolveAlert(alertId);
}
