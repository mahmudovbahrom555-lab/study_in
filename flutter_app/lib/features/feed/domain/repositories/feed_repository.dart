import '../entities/post.dart';

abstract class FeedRepository {
  Future<List<Post>> listPosts(
    String groupId, {
    int limit = 20,
    int offset = 0,
  });
  Future<Post> createPost(
    String groupId, {
    required String body,
    bool pinned = false,
  });
  Future<Post> updatePost(String postId, {String? body, bool? pinned});
  Future<void> deletePost(String postId);
  Future<Post> pinPost(String postId, {required bool pinned});

  /// null — снять свою реакцию.
  Future<void> setReaction(String postId, String? kind);
}
