import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/localization/l10n.dart';
import 'package:mamu_learn/features/grades/presentation/pages/grades_page.dart';
import 'package:mamu_learn/features/grades/presentation/widgets/add_grade_sheet.dart';
import 'package:mamu_learn/features/groups/domain/entities/group.dart';

final _students = [
  GroupMember(
    studentId: 's1',
    name: 'Bekzod',
    phone: '',
    joinedAt: DateTime(2026),
    paymentStatus: PaymentStatus.paid,
  ),
];

/// Лист открывается кнопкой — как в приложении; после закрытия кнопка снова видна.
Future<void> _open(WidgetTester tester, SubmitGrade onSubmit) async {
  await tester.pumpWidget(
    MaterialApp(
      locale: const Locale('ru'),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: Builder(
        builder: (context) => Scaffold(
          body: TextButton(
            onPressed: () => showModalBottomSheet<void>(
              context: context,
              isScrollControlled: true,
              builder: (_) => AddGradeSheet(students: _students, onSubmit: onSubmit),
            ),
            child: const Text('open'),
          ),
        ),
      ),
    ),
  );
  await tester.tap(find.text('open'));
  await tester.pumpAndSettle();
}

Future<void> _fill(WidgetTester tester, {required String value, String max = '5'}) async {
  await tester.tap(find.text('Ученик *'));
  await tester.pumpAndSettle();
  await tester.tap(find.text('Bekzod').last);
  await tester.pumpAndSettle();
  await tester.enterText(find.widgetWithText(TextFormField, 'За что *'), 'Диктант');
  await tester.enterText(find.widgetWithText(TextFormField, 'Макс.'), max);
  await tester.enterText(find.widgetWithText(TextFormField, 'Оценка *'), value);
  await tester.tap(find.text('Сохранить'));
  await tester.pumpAndSettle();
}

void main() {
  test('parseScore понимает запятую и точку', () {
    expect(parseScore('4,5'), 4.5);
    expect(parseScore(' 4.5 '), 4.5);
    expect(parseScore('0'), 0);
    expect(parseScore('abc'), isNull);
  });

  test('formatScore без лишних нулей, с разделителем языка', () {
    expect(formatScore(85, 'ru'), '85');
    expect(formatScore(4.5, 'ru'), '4,5');
    expect(formatScore(4.5, 'en'), '4.5');
  });

  testWidgets('0 и «4,5» принимаются, значение уходит числом', (tester) async {
    final sent = <double>[];
    await _open(tester, ({
      required studentId,
      required subject,
      required value,
      required maxValue,
      comment,
    }) async {
      sent.add(value);
      expect(studentId, 's1');
      expect(maxValue, 5);
      return null;
    });
    await _fill(tester, value: '4,5');
    expect(sent, [4.5]);
    expect(find.text('open'), findsOneWidget, reason: 'после успеха лист закрыт');

    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
    await _fill(tester, value: '0');
    expect(sent, [4.5, 0]);
  });

  testWidgets('балл выше максимума не отправляется', (tester) async {
    var calls = 0;
    await _open(tester, ({
      required studentId,
      required subject,
      required value,
      required maxValue,
      comment,
    }) async {
      calls++;
      return null;
    });
    await _fill(tester, value: '6');
    expect(calls, 0);
    expect(find.text('Не больше 5'), findsOneWidget);
  });

  testWidgets('ошибка сервера показывается, лист не закрывается', (tester) async {
    await _open(tester, ({
      required studentId,
      required subject,
      required value,
      required maxValue,
      comment,
    }) async {
      return 'Нет соединения с сервером';
    });
    await _fill(tester, value: '5');
    expect(find.text('Нет соединения с сервером'), findsOneWidget);
    expect(find.text('Сохранить'), findsOneWidget, reason: 'лист открыт, можно повторить');
  });
}
