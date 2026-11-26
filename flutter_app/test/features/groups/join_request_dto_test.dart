import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/config/app_config.dart';
import 'package:mamu_learn/features/groups/data/models/join_request_dto.dart';

void main() {
  test('ссылка-приглашение ведёт на карточку группы', () {
    expect(
      AppConfig.joinLink('AB3K9M2P'),
      'https://mamu.education/#/join/AB3K9M2P',
    );
  });

  test('заявка разбирается из ответа сервера', () {
    final r = joinRequestFromJson({
      'id': 'r1',
      'group_id': 'g1',
      'group_name': 'Математика',
      'status': 'pending',
    });
    expect(r.groupName, 'Математика');
    expect(r.status, 'pending');
  });

  test('карточка группы без предмета и описания', () {
    final g = groupPreviewFromJson({
      'id': 'g1',
      'name': 'Физика',
      'teacher_name': 'Анна',
      'member_count': 3,
    });
    expect(g.subject, isNull);
    expect(g.memberCount, 3);
  });
}
