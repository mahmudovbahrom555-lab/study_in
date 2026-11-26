import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/localization/l10n.dart';
import '../../../../core/router/routes.dart';
import '../../../../core/widgets/error_view.dart';
import '../../../../core/widgets/user_avatar.dart';
import '../../domain/entities/join_request.dart';
import '../providers/groups_provider.dart';

/// Поиск открытых групп по названию, предмету или @username репетитора.
/// Учеников поиск не находит. Группа из результата открывается карточкой,
/// где ученик отправляет заявку.
class GroupSearchPage extends ConsumerStatefulWidget {
  const GroupSearchPage({super.key});

  @override
  ConsumerState<GroupSearchPage> createState() => _GroupSearchPageState();
}

class _GroupSearchPageState extends ConsumerState<GroupSearchPage> {
  final _ctrl = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  void _search() {
    final q = _ctrl.text.trim();
    if (q.length < 2) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(context.l10n.minTwoChars)));
      return;
    }
    setState(() => _query = q);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.findGroup)),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: TextField(
              controller: _ctrl,
              autofocus: true,
              textInputAction: TextInputAction.search,
              onSubmitted: (_) => _search(),
              decoration: InputDecoration(
                hintText: context.l10n.findGroupHint,
                suffixIcon: IconButton(
                  icon: const Icon(Icons.search),
                  onPressed: _search,
                ),
              ),
            ),
          ),
          Expanded(child: _results()),
        ],
      ),
    );
  }

  Widget _results() {
    if (_query.isEmpty) {
      return _Hint(text: context.l10n.findGroupBody);
    }
    final async = ref.watch(groupSearchProvider(_query));
    return async.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => ErrorView(
        error: e,
        onRetry: () => ref.invalidate(groupSearchProvider(_query)),
      ),
      data: (groups) => groups.isEmpty
          ? _Hint(text: context.l10n.findGroupEmpty)
          : ListView.builder(
              itemCount: groups.length,
              itemBuilder: (context, i) => _ResultTile(group: groups[i]),
            ),
    );
  }
}

class _ResultTile extends StatelessWidget {
  const _ResultTile({required this.group});

  final GroupPreview group;

  @override
  Widget build(BuildContext context) {
    final username = group.teacherUsername;
    final tutor = username == null
        ? group.teacherName
        : '${group.teacherName} · @$username';
    final code = group.inviteCode;
    return ListTile(
      leading: UserAvatar(
        name: group.teacherName,
        avatarUrl: group.teacherAvatarUrl,
        avatarPreset: group.teacherAvatarPreset,
      ),
      title: Text(group.name),
      subtitle: Text(
        [if (group.subject != null) group.subject!, tutor].join('\n'),
      ),
      isThreeLine: group.subject != null,
      onTap: code == null ? null : () => context.push(Routes.joinGroup(code)),
    );
  }
}

class _Hint extends StatelessWidget {
  const _Hint({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Text(
          text,
          textAlign: TextAlign.center,
          style: TextStyle(color: Colors.grey.shade600),
        ),
      ),
    );
  }
}
