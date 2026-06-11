import '../domain/entities/post.dart';
import '../domain/repositories/feed_repository.dart';
import 'feed_api.dart';

class FeedRepositoryImpl implements FeedRepository {
  FeedRepositoryImpl(this._api, this._groupId);

  final FeedApi _api;
  final String _groupId;

  @override
  Future<List<Post>> listPosts(
    String groupId, {
    int limit = 20,
    int offset = 0,
  }) async {
    final dtos = await _api.listPosts(groupId, limit: limit, offset: offset);
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<Post> createPost(
    String groupId, {
    required String body,
    bool pinned = false,
  }) async {
    final dto = await _api.createPost(groupId, body: body, pinned: pinned);
    return dto.toDomain();
  }

  @override
  Future<Post> updatePost(String postId, {String? body, bool? pinned}) async {
    final dto = await _api.updatePost(postId, _groupId, body: body, pinned: pinned);
    return dto.toDomain();
  }

  @override
  Future<void> deletePost(String postId) => _api.deletePost(postId, _groupId);

  @override
  Future<Post> pinPost(String postId, {required bool pinned}) async {
    final dto = await _api.pinPost(postId, _groupId, pinned: pinned);
    return dto.toDomain();
  }
}
