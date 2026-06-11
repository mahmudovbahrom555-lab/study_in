import 'package:flutter/foundation.dart';

enum PaymentStatus {
  trial,
  pending,
  paid;

  static PaymentStatus fromJson(String v) => switch (v) {
        'paid' => paid,
        'pending' => pending,
        _ => trial,
      };

  String toJson() => name;
}

@immutable
class Group {
  const Group({
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

  Group copyWith({
    String? name,
    String? subject,
    String? description,
    bool? isArchived,
    int? memberCount,
  }) =>
      Group(
        id: id,
        teacherId: teacherId,
        name: name ?? this.name,
        subject: subject ?? this.subject,
        description: description ?? this.description,
        inviteCode: inviteCode,
        isArchived: isArchived ?? this.isArchived,
        memberCount: memberCount ?? this.memberCount,
        createdAt: createdAt,
      );
}

@immutable
class GroupMember {
  const GroupMember({
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
}
