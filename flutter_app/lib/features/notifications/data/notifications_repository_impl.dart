import '../domain/entities/notification.dart';
import '../domain/repositories/notifications_repository.dart';
import 'notifications_api.dart';

class NotificationsRepositoryImpl implements NotificationsRepository {
  NotificationsRepositoryImpl(this._api);

  final NotificationsApi _api;

  @override
  Future<List<AppNotification>> listNotifications() async {
    final dtos = await _api.listNotifications();
    return dtos.map((d) => d.toDomain()).toList();
  }

  @override
  Future<void> markRead(String id) => _api.markRead(id);

  @override
  Future<void> markAllRead() => _api.markAllRead();
}
