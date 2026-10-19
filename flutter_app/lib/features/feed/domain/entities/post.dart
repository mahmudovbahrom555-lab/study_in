import 'package:flutter/foundation.dart';

@immutable
class Post {
  const Post({
    required this.id,
    required this.groupId,
    required this.authorId,
    required this.authorName,
    required this.body,
    required this.pinned,
    required this.createdAt,
    required this.updatedAt,
  });

  final String id;
  final String groupId;
  final String authorId;
  final String authorName;
  final String body;
  final bool pinned;
  final DateTime createdAt;
  final DateTime updatedAt;

  Post copyWith({
    String? body,
    bool? pinned,
  }) =>
      Post(
        id: id,
        groupId: groupId,
        authorId: authorId,
        authorName: authorName,
        body: body ?? this.body,
        pinned: pinned ?? this.pinned,
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}
