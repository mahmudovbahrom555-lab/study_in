import '../../domain/entities/notification.dart';

class NotificationDto {
  factory NotificationDto.fromJson(Map<String, dynamic> json) =>
      NotificationDto(
        id: json['id'] as String,
        title: json['title'] as String,
        body: json['body'] as String,
        isRead: json['is_read'] as bool? ?? false,
        createdAt: DateTime.parse(json['created_at'] as String),
        type: json['type'] as String?,
      );

  const NotificationDto({
    required this.id,
    required this.title,
    required this.body,
    required this.isRead,
    required this.createdAt,
    this.type,
  });

  final String id;
  final String title;
  final String body;
  final String? type;
  final bool isRead;
  final DateTime createdAt;

  AppNotification toDomain() => AppNotification(
        id: id,
        title: title,
        body: body,
        type: type,
        isRead: isRead,
        createdAt: createdAt,
      );
}
