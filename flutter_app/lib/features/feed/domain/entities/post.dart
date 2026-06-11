import 'package:flutter/foundation.dart';

@immutable
class PostAttachment {
  const PostAttachment({
    required this.id,
    required this.filename,
    required this.mimeType,
    required this.sizeBytes,
    required this.url,
  });

  final String id;
  final String filename;
  final String mimeType;
  final int sizeBytes;
  final String url;
}

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
    this.attachments = const [],
  });

  final String id;
  final String groupId;
  final String authorId;
  final String authorName;
  final String body;
  final bool pinned;
  final List<PostAttachment> attachments;
  final DateTime createdAt;
  final DateTime updatedAt;

  Post copyWith({
    String? body,
    bool? pinned,
    List<PostAttachment>? attachments,
  }) =>
      Post(
        id: id,
        groupId: groupId,
        authorId: authorId,
        authorName: authorName,
        body: body ?? this.body,
        pinned: pinned ?? this.pinned,
        attachments: attachments ?? this.attachments,
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}
