import '../../domain/entities/post.dart';

class PostDto {
  factory PostDto.fromJson(Map<String, dynamic> json) => PostDto(
        id: json['id'] as String,
        groupId: json['group_id'] as String,
        authorId: json['author_id'] as String,
        authorName: json['author_name'] as String? ?? '',
        body: json['body'] as String,
        pinned: json['pinned'] as bool? ?? false,
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

  Post toDomain() => Post(
        id: id,
        groupId: groupId,
        authorId: authorId,
        authorName: authorName,
        body: body,
        pinned: pinned,
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}
