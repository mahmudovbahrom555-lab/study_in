import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../auth/presentation/providers/auth_provider.dart';
import '../providers/assignments_provider.dart';
import 'review_submissions_page.dart';
import 'submit_work_page.dart';

/// Задание по ссылке: репетитору — проверка работ, ученику — сдача.
/// Задание берём из уже загруженного списка группы.
class AssignmentPage extends ConsumerStatefulWidget {
  const AssignmentPage({
    super.key,
    required this.groupId,
    required this.assignmentId,
  });

  final String groupId;
  final String assignmentId;

  @override
  ConsumerState<AssignmentPage> createState() => _AssignmentPageState();
}

class _AssignmentPageState extends ConsumerState<AssignmentPage> {
  @override
  void initState() {
    super.initState();
    final state = ref.read(assignmentsProvider(widget.groupId));
    if (state.assignments.isEmpty && !state.isLoading) {
      Future.microtask(
        () => ref.read(assignmentsProvider(widget.groupId).notifier).load(),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final a = ref
        .watch(assignmentsProvider(widget.groupId))
        .assignments
        .where((a) => a.id == widget.assignmentId)
        .firstOrNull;
    if (a == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    final isTeacher = ref.watch(authProvider).user?.role == 'teacher';
    return isTeacher
        ? ReviewSubmissionsPage(assignment: a)
        : SubmitWorkPage(assignment: a);
  }
}
