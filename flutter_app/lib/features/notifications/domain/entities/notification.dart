import 'package:flutter/foundation.dart';

@immutable
class AppNotification {
  const AppNotification({
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

  AppNotification copyWith({
    bool? isRead,
  }) =>
      AppNotification(
        id: id,
        title: title,
        body: body,
        type: type,
        isRead: isRead ?? this.isRead,
        createdAt: createdAt,
      );
}
