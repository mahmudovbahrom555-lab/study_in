import 'package:flutter/foundation.dart';

@immutable
class Post {
  const Post({
    required this.id,
    required this.groupId,
    required this.authorId,
    required this.authorName,
    required this.body,
    required this.pinned,
    required this.createdAt,
    required this.updatedAt,
    this.reactions = const PostReactions(),
  });

  final String id;
  final String groupId;
  final String authorId;
  final String authorName;
  final String body;
  final bool pinned;
  final DateTime createdAt;
  final DateTime updatedAt;
  final PostReactions reactions;

  Post copyWith({String? body, bool? pinned, PostReactions? reactions}) => Post(
    id: id,
    groupId: groupId,
    authorId: authorId,
    authorName: authorName,
    body: body ?? this.body,
    pinned: pinned ?? this.pinned,
    createdAt: createdAt,
    updatedAt: updatedAt,
    reactions: reactions ?? this.reactions,
  );
}

/// Реакции учеников: «Понятно» и «Есть вопрос». Имена приходят только
/// репетитору класса.
@immutable
class PostReactions {
  const PostReactions({
    this.understood = 0,
    this.question = 0,
    this.mine,
    this.understoodBy = const [],
    this.questionBy = const [],
  });

  static const understoodKind = 'understood';
  static const questionKind = 'question';

  final int understood;
  final int question;

  /// Своя реакция ученика; null — не ставил.
  final String? mine;
  final List<String> understoodBy;
  final List<String> questionBy;

  /// Нажатие ученика: та же реакция — снять, другая — заменить.
  PostReactions toggled(String kind) {
    var understood = this.understood;
    var question = this.question;
    if (mine == understoodKind) understood--;
    if (mine == questionKind) question--;
    final next = mine == kind ? null : kind;
    if (next == understoodKind) understood++;
    if (next == questionKind) question++;
    return PostReactions(
      understood: understood,
      question: question,
      mine: next,
      understoodBy: understoodBy,
      questionBy: questionBy,
    );
  }
}
