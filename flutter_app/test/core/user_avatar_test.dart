import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/widgets/user_avatar.dart';

Widget _wrap(Widget child) => MaterialApp(home: Scaffold(body: child));

void main() {
  testWidgets('животное показывается эмодзи', (tester) async {
    await tester.pumpWidget(
      _wrap(const UserAvatar(name: 'Ali', avatarPreset: 'fox')),
    );
    expect(find.text('🦊'), findsOneWidget);
  });

  testWidgets('нет аватара — первая буква имени', (tester) async {
    await tester.pumpWidget(_wrap(const UserAvatar(name: 'asila')));
    expect(find.text('A'), findsOneWidget);
  });

  testWidgets('неизвестное животное — тоже буква, без падения', (tester) async {
    await tester.pumpWidget(
      _wrap(const UserAvatar(name: 'Vali', avatarPreset: 'dragon')),
    );
    expect(find.text('V'), findsOneWidget);
  });

  test('набор животных совпадает с серверным', () {
    expect(avatarPresets.keys, [
      'fox', 'panda', 'koala', 'lion', 'tiger', 'frog', //
      'monkey', 'penguin', 'bear', 'rabbit', 'cat', 'dog',
    ]);
  });
}
