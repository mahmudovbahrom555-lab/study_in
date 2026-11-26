import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../auth/presentation/providers/auth_provider.dart';
import '../providers/feed_provider.dart';
import '../widgets/create_post_sheet.dart';
import '../widgets/post_card.dart';
import '../../../../core/widgets/error_view.dart';
import '../../../../core/localization/l10n.dart';

class FeedPage extends ConsumerStatefulWidget {
  const FeedPage({super.key, required this.groupId});

  final String groupId;

  @override
  ConsumerState<FeedPage> createState() => _FeedPageState();
}

class _FeedPageState extends ConsumerState<FeedPage> {
  final _scrollCtrl = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollCtrl.addListener(_onScroll);
    // Лента живёт всю сессию — при повторном открытии берём свежие объявления
    // (при первом её уже загружает сам провайдер).
    Future.microtask(() {
      final notifier = ref.read(feedProvider(widget.groupId).notifier);
      if (!ref.read(feedProvider(widget.groupId)).isLoading) notifier.load();
    });
  }

  @override
  void dispose() {
    _scrollCtrl.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollCtrl.position.pixels >=
        _scrollCtrl.position.maxScrollExtent - 200) {
      ref.read(feedProvider(widget.groupId).notifier).loadMore();
    }
  }

  @override
  Widget build(BuildContext context) {
    final feedState = ref.watch(feedProvider(widget.groupId));
    final authUser = ref.watch(authProvider).user;
    final isTeacher = authUser?.role == 'teacher';

    return Scaffold(
      appBar: AppBar(title: Text(context.l10n.tabFeed)),
      floatingActionButton: isTeacher
          ? FloatingActionButton(
              tooltip: context.l10n.newPostTooltip,
              onPressed: () => _showCreateSheet(context),
              child: const Icon(LucideIcons.squarePen),
            )
          : null,
      body: _buildBody(feedState, authUser?.id ?? '', isTeacher),
    );
  }

  Future<void> _react(String postId, String kind) async {
    final messenger = ScaffoldMessenger.of(context);
    final error = await ref
        .read(feedProvider(widget.groupId).notifier)
        .toggleReaction(postId, kind);
    if (error != null) messenger.showSnackBar(SnackBar(content: Text(error)));
  }

  Widget _buildBody(dynamic feedState, String userId, bool isTeacher) {
    final isStudent = ref.watch(authProvider).user?.isStudent ?? false;
    if (feedState.isLoading && feedState.posts.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (feedState.error != null && feedState.posts.isEmpty) {
      return ErrorView(
        message: feedState.error as String,
        onRetry: () => ref.read(feedProvider(widget.groupId).notifier).load(),
      );
    }
    if (feedState.posts.isEmpty) {
      return Center(child: Text(context.l10n.noPosts));
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(feedProvider(widget.groupId).notifier).load(),
      child: ListView.separated(
        controller: _scrollCtrl,
        padding: const EdgeInsets.all(16),
        itemCount: feedState.posts.length + (feedState.hasMore ? 1 : 0),
        separatorBuilder: (_, __) => const SizedBox(height: 8),
        itemBuilder: (context, i) {
          if (i == feedState.posts.length) {
            return const Center(child: CircularProgressIndicator());
          }
          final post = feedState.posts[i];
          return PostCard(
            post: post,
            isAuthor: post.authorId == userId,
            isTeacher: isTeacher,
            onDelete: isTeacher || post.authorId == userId
                ? () => ref
                      .read(feedProvider(widget.groupId).notifier)
                      .deletePost(post.id)
                : null,
            onPin: isTeacher
                ? () => ref
                      .read(feedProvider(widget.groupId).notifier)
                      .pinPost(post.id, pinned: !post.pinned)
                : null,
            onReact: isStudent ? (kind) => _react(post.id, kind) : null,
          );
        },
      ),
    );
  }

  void _showCreateSheet(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (_) => CreatePostSheet(groupId: widget.groupId),
    );
  }
}
