import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/ai_api.dart';
import '../../data/ai_repository_impl.dart';
import '../../domain/entities/class_insights.dart';
import '../../domain/repositories/ai_repository.dart';

final aiRepositoryProvider = Provider<AiRepository>((ref) {
  final dio = ref.watch(dioProvider);
  return AiRepositoryImpl(AiApi(dio));
});

// ─── ClassInsights ────────────────────────────────────────────────────────────

final classInsightsProvider =
    FutureProvider.family<ClassInsights, String>((ref, groupId) {
  return ref.watch(aiRepositoryProvider).classInsights(groupId);
});

// ─── StudentProgress ──────────────────────────────────────────────────────────

typedef _ProgressArgs = ({String groupId, String studentId});

final studentProgressProvider =
    FutureProvider.family<StudentProgress, _ProgressArgs>((ref, args) {
  return ref
      .watch(aiRepositoryProvider)
      .studentProgress(args.groupId, args.studentId);
});

// ─── Acceptance rate ──────────────────────────────────────────────────────────

final acceptanceRateProvider =
    FutureProvider<({double rate, int total})>((ref) {
  return ref.watch(aiRepositoryProvider).myAcceptanceRate();
});

// ─── Generation stats (TAR + session counts) ─────────────────────────────────

final generationStatsProvider =
    FutureProvider<TeacherGenerationStats>((ref) {
  return ref.watch(aiRepositoryProvider).generationStats();
});
