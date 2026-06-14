import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/reports_api.dart';
import '../../data/reports_repository_impl.dart';
import '../../domain/entities/report.dart';
import '../../domain/repositories/reports_repository.dart';

final reportsRepositoryProvider = Provider<ReportsRepository>((ref) {
  final dio = ref.watch(dioProvider);
  return ReportsRepositoryImpl(ReportsApi(dio));
});

// ─── Parent ROI ───────────────────────────────────────────────────────────────

typedef _RoiArgs = ({String studentId, String groupId});

final parentRoiProvider =
    FutureProvider.family<ParentReport, _RoiArgs>((ref, args) {
  return ref
      .watch(reportsRepositoryProvider)
      .getParentRoi(args.studentId, args.groupId);
});

// ─── Owner Risk Alerts ────────────────────────────────────────────────────────

class RiskAlertsNotifier extends StateNotifier<AsyncValue<List<RiskAlert>>> {
  RiskAlertsNotifier(this._repo) : super(const AsyncValue.loading()) {
    load();
  }

  final ReportsRepository _repo;

  Future<void> load({String status = ''}) async {
    state = const AsyncValue.loading();
    state = await AsyncValue.guard(() => _repo.getRiskAlerts(status: status));
  }

  Future<void> refresh() async {
    state = const AsyncValue.loading();
    state = await AsyncValue.guard(() => _repo.refreshRiskAlerts());
  }

  Future<void> resolve(String alertId) async {
    await _repo.resolveAlert(alertId);
    final current = state.valueOrNull ?? [];
    state = AsyncValue.data(
      current
          .map((a) => a.id == alertId
              ? RiskAlert(
                  id: a.id,
                  studentId: a.studentId,
                  studentName: a.studentName,
                  groupId: a.groupId,
                  groupName: a.groupName,
                  riskLevel: a.riskLevel,
                  triggerReason: a.triggerReason,
                  recommendation: a.recommendation,
                  status: 'resolved',
                  createdAt: a.createdAt,
                )
              : a,)
          .toList(),
    );
  }
}

final riskAlertsProvider =
    StateNotifierProvider<RiskAlertsNotifier, AsyncValue<List<RiskAlert>>>(
  (ref) => RiskAlertsNotifier(ref.watch(reportsRepositoryProvider)),
);
