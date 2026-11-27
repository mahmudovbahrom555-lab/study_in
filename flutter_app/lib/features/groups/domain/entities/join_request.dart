import 'package:flutter/foundation.dart';

/// Своя заявка ученика в группу.
@immutable
class JoinRequest {
  const JoinRequest({
    required this.id,
    required this.groupId,
    required this.groupName,
    required this.status,
    this.hasParent = false,
  });

  final String id;
  final String groupId;
  final String groupName;

  /// 'parent_pending' (ребёнок младше 13: ждёт родителя), 'pending',
  /// 'accepted' или 'declined'.
  final String status;

  /// У заявки ребёнка: есть ли привязанный родитель, которого уведомили.
  final bool hasParent;

  bool get waitsForParent => status == 'parent_pending';
}

/// Заявка в списке репетитора.
@immutable
class JoinRequestItem {
  const JoinRequestItem({
    required this.id,
    required this.studentName,
    this.avatarUrl,
    this.avatarPreset,
    this.parentName,
  });

  final String id;
  final String studentName;
  final String? avatarUrl;
  final String? avatarPreset;

  /// Родитель, разрешивший заявку ребёнка младше 13.
  final String? parentName;
}

/// Заявка ребёнка, которая ждёт согласия родителя.
@immutable
class ParentConsent {
  const ParentConsent({
    required this.id,
    required this.studentName,
    required this.groupName,
    required this.teacherName,
    this.subject,
  });

  final String id;
  final String studentName;
  final String groupName;
  final String teacherName;
  final String? subject;
}

/// Карточка группы по ссылке — что видит ученик до заявки. В поиске
/// открытых групп у неё есть ещё код (для заявки) и @username репетитора.
@immutable
class GroupPreview {
  const GroupPreview({
    required this.id,
    required this.name,
    required this.teacherName,
    required this.memberCount,
    this.subject,
    this.description,
    this.teacherAvatarUrl,
    this.teacherAvatarPreset,
    this.inviteCode,
    this.teacherUsername,
  });

  final String id;
  final String name;
  final String teacherName;
  final int memberCount;
  final String? subject;
  final String? description;
  final String? teacherAvatarUrl;
  final String? teacherAvatarPreset;
  final String? inviteCode;
  final String? teacherUsername;
}
