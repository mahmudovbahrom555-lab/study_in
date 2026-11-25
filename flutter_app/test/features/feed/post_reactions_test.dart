import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/features/feed/data/models/post_dto.dart';
import 'package:mamu_learn/features/feed/domain/entities/post.dart';

void main() {
  const understood = PostReactions.understoodKind;
  const question = PostReactions.questionKind;

  test('ученик ставит, меняет и снимает реакцию', () {
    const start = PostReactions(understood: 2, question: 1);

    final put = start.toggled(understood);
    expect((put.understood, put.question, put.mine), (3, 1, understood));

    final changed = put.toggled(question);
    expect((changed.understood, changed.question, changed.mine), (
      2,
      2,
      question,
    ));

    final cleared = changed.toggled(question);
    expect((cleared.understood, cleared.question, cleared.mine), (2, 1, null));
  });

  test('ответ сервера: счётчики, своя реакция, имена для репетитора', () {
    final post = PostDto.fromJson({
      'id': 'p1',
      'group_id': 'g1',
      'author_id': 't1',
      'author_name': 'Murod',
      'body': 'Exam on Friday',
      'pinned': false,
      'created_at': '2026-11-25T10:00:00Z',
      'updated_at': '2026-11-25T10:00:00Z',
      'reactions': {
        'understood': 1,
        'question': 1,
        'question_by': ['Asila'],
      },
    }).toDomain();

    expect(post.reactions.understood, 1);
    expect(post.reactions.question, 1);
    expect(post.reactions.mine, isNull);
    expect(post.reactions.questionBy, ['Asila']);
    expect(post.reactions.understoodBy, isEmpty);
  });

  test('старый ответ без reactions — нули', () {
    final post = PostDto.fromJson({
      'id': 'p1',
      'group_id': 'g1',
      'author_id': 't1',
      'body': 'x',
      'created_at': '2026-11-25T10:00:00Z',
      'updated_at': '2026-11-25T10:00:00Z',
    }).toDomain();

    expect(post.reactions.understood, 0);
    expect(post.reactions.mine, isNull);
  });
}
