import '../../domain/entities/join_request.dart';

JoinRequest joinRequestFromJson(Map<String, dynamic> json) => JoinRequest(
  id: json['id'] as String,
  groupId: json['group_id'] as String,
  groupName: json['group_name'] as String? ?? '',
  status: json['status'] as String? ?? 'pending',
  hasParent: json['has_parent'] as bool? ?? false,
);

JoinRequestItem joinRequestItemFromJson(Map<String, dynamic> json) =>
    JoinRequestItem(
      id: json['id'] as String,
      studentName: json['student_name'] as String? ?? '',
      avatarUrl: json['avatar_url'] as String?,
      avatarPreset: json['avatar_preset'] as String?,
      parentName: json['approved_by_parent_name'] as String?,
    );

ParentConsent parentConsentFromJson(Map<String, dynamic> json) => ParentConsent(
  id: json['id'] as String,
  studentName: json['student_name'] as String? ?? '',
  groupName: json['group_name'] as String? ?? '',
  teacherName: json['teacher_name'] as String? ?? '',
  subject: json['subject'] as String?,
);

GroupPreview groupPreviewFromJson(Map<String, dynamic> json) => GroupPreview(
  id: json['id'] as String,
  name: json['name'] as String,
  teacherName: json['teacher_name'] as String? ?? '',
  memberCount: json['member_count'] as int? ?? 0,
  subject: json['subject'] as String?,
  description: json['description'] as String?,
  teacherAvatarUrl: json['teacher_avatar_url'] as String?,
  teacherAvatarPreset: json['teacher_avatar_preset'] as String?,
  inviteCode: json['invite_code'] as String?,
  teacherUsername: json['teacher_username'] as String?,
);
