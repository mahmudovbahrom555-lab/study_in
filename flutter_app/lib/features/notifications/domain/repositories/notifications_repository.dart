import '../entities/notification.dart';

abstract class NotificationsRepository {
  Future<List<AppNotification>> listNotifications();
  Future<void> markRead(String id);
  Future<void> markAllRead();
}
