import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/feed_api.dart';
import '../../data/feed_repository_impl.dart';
import '../../domain/entities/post.dart';
import '../../domain/repositories/feed_repository.dart';

// Per-group feed provider family.
final feedRepositoryProvider =
    Provider.family<FeedRepository, String>((ref, groupId) {
  final dio = ref.watch(dioProvider);
  return FeedRepositoryImpl(FeedApi(dio), groupId);
});

// ─── State ───────────────────────────────────────────────────────────────────

class FeedState {
  const FeedState({
    this.posts = const [],
    this.isLoading = false,
    this.hasMore = true,
    this.error,
  });

  final List<Post> posts;
  final bool isLoading;
  final bool hasMore;
  final String? error;

  FeedState copyWith({
    List<Post>? posts,
    bool? isLoading,
    bool? hasMore,
    String? error,
    bool clearError = false,
  }) =>
      FeedState(
        posts: posts ?? this.posts,
        isLoading: isLoading ?? this.isLoading,
        hasMore: hasMore ?? this.hasMore,
        error: clearError ? null : (error ?? this.error),
      );
}

// ─── Notifier ─────────────────────────────────────────────────────────────────

class FeedNotifier extends StateNotifier<FeedState> {
  FeedNotifier(this._repo, this._groupId) : super(const FeedState()) {
    load();
  }

  final FeedRepository _repo;
  final String _groupId;

  static const _pageSize = 20;

  Future<void> load() async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final posts = await _repo.listPosts(_groupId, limit: _pageSize, offset: 0);
      state = state.copyWith(
        isLoading: false,
        posts: posts,
        hasMore: posts.length == _pageSize,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> loadMore() async {
    if (state.isLoading || !state.hasMore) return;
    state = state.copyWith(isLoading: true);
    try {
      final more = await _repo.listPosts(
        _groupId,
        limit: _pageSize,
        offset: state.posts.length,
      );
      state = state.copyWith(
        isLoading: false,
        posts: [...state.posts, ...more],
        hasMore: more.length == _pageSize,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> createPost(String body, {bool pinned = false}) async {
    try {
      final p = await _repo.createPost(_groupId, body: body, pinned: pinned);
      state = state.copyWith(posts: [p, ...state.posts]);
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> deletePost(String postId) async {
    try {
      await _repo.deletePost(postId);
      state = state.copyWith(
        posts: state.posts.where((p) => p.id != postId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> pinPost(String postId, {required bool pinned}) async {
    try {
      final updated = await _repo.pinPost(postId, pinned: pinned);
      state = state.copyWith(
        posts: state.posts.map((p) => p.id == postId ? updated : p).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

final feedProvider =
    StateNotifierProvider.family<FeedNotifier, FeedState, String>(
  (ref, groupId) =>
      FeedNotifier(ref.watch(feedRepositoryProvider(groupId)), groupId),
);
