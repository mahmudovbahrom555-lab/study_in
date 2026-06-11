import '../../domain/entities/post.dart';

class PostAttachmentDto {
  factory PostAttachmentDto.fromJson(Map<String, dynamic> json) =>
      PostAttachmentDto(
        id: json['id'] as String,
        filename: json['filename'] as String,
        mimeType: json['mime_type'] as String,
        sizeBytes: json['size_bytes'] as int? ?? 0,
        url: json['url'] as String? ?? '',
      );

  const PostAttachmentDto({
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

  PostAttachment toDomain() => PostAttachment(
        id: id,
        filename: filename,
        mimeType: mimeType,
        sizeBytes: sizeBytes,
        url: url,
      );
}

class PostDto {
  factory PostDto.fromJson(Map<String, dynamic> json) => PostDto(
        id: json['id'] as String,
        groupId: json['group_id'] as String,
        authorId: json['author_id'] as String,
        authorName: json['author_name'] as String? ?? '',
        body: json['body'] as String,
        pinned: json['pinned'] as bool? ?? false,
        attachments: (json['attachments'] as List<dynamic>?)
                ?.map((e) =>
                    PostAttachmentDto.fromJson(e as Map<String, dynamic>))
                .toList() ??
            [],
        createdAt: DateTime.parse(json['created_at'] as String),
        updatedAt: DateTime.parse(json['updated_at'] as String),
      );

  const PostDto({
    required this.id,
    required this.groupId,
    required this.authorId,
    required this.authorName,
    required this.body,
    required this.pinned,
    required this.attachments,
    required this.createdAt,
    required this.updatedAt,
  });

  final String id;
  final String groupId;
  final String authorId;
  final String authorName;
  final String body;
  final bool pinned;
  final List<PostAttachmentDto> attachments;
  final DateTime createdAt;
  final DateTime updatedAt;

  Post toDomain() => Post(
        id: id,
        groupId: groupId,
        authorId: authorId,
        authorName: authorName,
        body: body,
        pinned: pinned,
        attachments: attachments.map((a) => a.toDomain()).toList(),
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}
