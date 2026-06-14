import 'package:dio/dio.dart';

import 'models/notification_dto.dart';

class NotificationsApi {
  NotificationsApi(this._dio);

  final Dio _dio;

  Future<List<NotificationDto>> listNotifications() async {
    final resp = await _dio.get<Map<String, dynamic>>('/notifications');
    final list = resp.data!['data'] as List<dynamic>;
    return list
        .map((e) => NotificationDto.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> markRead(String id) =>
      _dio.post<void>('/notifications/$id/read');

  Future<void> markAllRead() => _dio.post<void>('/notifications/read-all');
}
