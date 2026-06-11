import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../auth/presentation/providers/auth_provider.dart';
import '../providers/feed_provider.dart';
import '../widgets/post_card.dart';
import '../widgets/create_post_sheet.dart';

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
      appBar: AppBar(title: const Text('Лента')),
      floatingActionButton: isTeacher
          ? FloatingActionButton(
              onPressed: () => _showCreateSheet(context),
              child: const Icon(Icons.edit),
            )
          : null,
      body: _buildBody(feedState, authUser?.id ?? '', isTeacher),
    );
  }

  Widget _buildBody(dynamic feedState, String userId, bool isTeacher) {
    if (feedState.isLoading && feedState.posts.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (feedState.error != null && feedState.posts.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(feedState.error.toString(),
                style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: () =>
                  ref.read(feedProvider(widget.groupId).notifier).load(),
              child: const Text('Повторить'),
            ),
          ],
        ),
      );
    }
    if (feedState.posts.isEmpty) {
      return const Center(child: Text('Нет публикаций'));
    }

    return RefreshIndicator(
      onRefresh: () =>
          ref.read(feedProvider(widget.groupId).notifier).load(),
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
