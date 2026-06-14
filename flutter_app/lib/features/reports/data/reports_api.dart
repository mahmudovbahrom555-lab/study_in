import 'package:dio/dio.dart';

import 'models/report_dto.dart';

class ReportsApi {
  ReportsApi(this._dio);

  final Dio _dio;

  Future<ParentReportDto> getParentRoi(String studentId, String groupId) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/parent/children/$studentId/groups/$groupId/roi',
    );
    return ParentReportDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<List<RiskAlertDto>> getRiskAlerts({String status = ''}) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/owner/risk-alerts',
      queryParameters: status.isNotEmpty ? {'status': status} : null,
    );
    return (resp.data!['data'] as List<dynamic>)
        .map((e) => RiskAlertDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<RiskAlertDto>> refreshRiskAlerts() async {
    final resp = await _dio.post<Map<String, dynamic>>('/owner/risk-alerts/refresh');
    return (resp.data!['data'] as List<dynamic>)
        .map((e) => RiskAlertDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> resolveAlert(String alertId) =>
      _dio.patch<void>('/owner/risk-alerts/$alertId/resolve');
}
