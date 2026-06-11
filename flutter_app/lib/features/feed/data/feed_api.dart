import 'package:dio/dio.dart';

import 'models/post_dto.dart';

class FeedApi {
  FeedApi(this._dio);

  final Dio _dio;

  Future<List<PostDto>> listPosts(
    String groupId, {
    int limit = 20,
    int offset = 0,
  }) async {
    final resp = await _dio.get<Map<String, dynamic>>(
      '/groups/$groupId/feed',
      queryParameters: {'limit': limit, 'offset': offset},
    );
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => PostDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<PostDto> createPost(
    String groupId, {
    required String body,
    bool pinned = false,
  }) async {
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/feed',
      data: {'body': body, 'pinned': pinned},
    );
    return PostDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<PostDto> updatePost(
    String postId,
    String groupId, {
    String? body,
    bool? pinned,
  }) async {
    final resp = await _dio.patch<Map<String, dynamic>>(
      '/groups/$groupId/feed/$postId',
      data: {
        if (body != null) 'body': body,
        if (pinned != null) 'pinned': pinned,
      },
    );
    return PostDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }

  Future<void> deletePost(String postId, String groupId) =>
      _dio.delete<void>('/groups/$groupId/feed/$postId');

  Future<PostDto> pinPost(
    String postId,
    String groupId, {
    required bool pinned,
  }) async {
    final endpoint = pinned ? 'pin' : 'unpin';
    final resp = await _dio.post<Map<String, dynamic>>(
      '/groups/$groupId/feed/$postId/$endpoint',
    );
    return PostDto.fromJson(resp.data!['data'] as Map<String, dynamic>);
  }
}
