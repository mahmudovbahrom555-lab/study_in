import '../../domain/entities/group.dart';

class GroupDto {
  factory GroupDto.fromJson(Map<String, dynamic> json) => GroupDto(
        id: json['id'] as String,
        teacherId: json['teacher_id'] as String,
        name: json['name'] as String,
        subject: json['subject'] as String?,
        description: json['description'] as String?,
        inviteCode: json['invite_code'] as String,
        isArchived: json['is_archived'] as bool? ?? false,
        memberCount: json['member_count'] as int? ?? 0,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  const GroupDto({
    required this.id,
    required this.teacherId,
    required this.name,
    required this.inviteCode,
    required this.isArchived,
    required this.createdAt,
    this.subject,
    this.description,
    this.memberCount = 0,
  });

  final String id;
  final String teacherId;
  final String name;
  final String? subject;
  final String? description;
  final String inviteCode;
  final bool isArchived;
  final int memberCount;
  final DateTime createdAt;

  Group toDomain() => Group(
        id: id,
        teacherId: teacherId,
        name: name,
        subject: subject,
        description: description,
        inviteCode: inviteCode,
        isArchived: isArchived,
        memberCount: memberCount,
        createdAt: createdAt,
      );
}

class GroupMemberDto {
  factory GroupMemberDto.fromJson(Map<String, dynamic> json) => GroupMemberDto(
        studentId: json['student_id'] as String,
        name: json['name'] as String,
        phone: json['phone'] as String,
        avatarUrl: json['avatar_url'] as String?,
        joinedAt: DateTime.parse(json['joined_at'] as String),
        paymentStatus: PaymentStatus.fromJson(json['payment_status'] as String),
      );

  const GroupMemberDto({
    required this.studentId,
    required this.name,
    required this.phone,
    required this.joinedAt,
    required this.paymentStatus,
    this.avatarUrl,
  });

  final String studentId;
  final String name;
  final String phone;
  final String? avatarUrl;
  final DateTime joinedAt;
  final PaymentStatus paymentStatus;

  GroupMember toDomain() => GroupMember(
        studentId: studentId,
        name: name,
        phone: phone,
        avatarUrl: avatarUrl,
        joinedAt: joinedAt,
        paymentStatus: paymentStatus,
      );
}
