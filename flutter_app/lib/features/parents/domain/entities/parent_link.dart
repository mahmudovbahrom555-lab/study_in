import 'package:flutter/foundation.dart';

@immutable
class ParentLink {
  const ParentLink({
    required this.id,
    required this.parentId,
    required this.studentId,
    required this.studentName,
    required this.createdAt,
  });

  final String id;
  final String parentId;
  final String studentId;

  /// Пусто только в ответе на привязку по коду — список детей приходит с именем.
  final String studentName;
  final DateTime createdAt;
}
