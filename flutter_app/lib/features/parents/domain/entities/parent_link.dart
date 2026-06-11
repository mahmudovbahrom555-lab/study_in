import 'package:flutter/foundation.dart';

@immutable
class ParentLink {
  const ParentLink({
    required this.id,
    required this.parentId,
    required this.studentId,
    required this.createdAt,
  });

  final String id;
  final String parentId;
  final String studentId;
  final DateTime createdAt;
}
