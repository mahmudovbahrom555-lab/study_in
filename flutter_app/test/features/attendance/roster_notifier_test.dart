import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/features/attendance/domain/entities/attendance.dart';
import 'package:mamu_learn/features/attendance/domain/repositories/attendance_repository.dart';
import 'package:mamu_learn/features/attendance/presentation/providers/attendance_provider.dart';

/// Репозиторий в памяти; [fail] — следующий вызов записи падает как без сети.
class _FakeRepo implements AttendanceRepository {
  final marks = <String, AttendanceStatus>{}; // studentId → статус за день
  bool fail = false;
  Completer<void>? gate; // держит mark(), чтобы проверить повторный тап
  int markCalls = 0;

  static const names = {'s1': 'Aziza', 's2': 'Bekzod'};

  void _maybeFail() {
    if (fail) {
      throw DioException(
        requestOptions: RequestOptions(),
        type: DioExceptionType.connectionError,
      );
    }
  }

  Attendance _rec(String sid) => Attendance(
        id: 'a-$sid',
        groupId: 'g',
        studentId: sid,
        teacherId: 't',
        lessonDate: '2026-10-14',
        status: marks[sid]!,
        createdAt: DateTime(2026),
        updatedAt: DateTime(2026),
      );

  @override
  Future<List<RosterEntry>> roster(String groupId, String date) async => [
        for (final e in names.entries)
          RosterEntry(
            studentId: e.key,
            studentName: e.value,
            attendanceId: marks.containsKey(e.key) ? 'a-${e.key}' : null,
            status: marks[e.key],
          ),
      ];

  @override
  Future<Attendance> mark({
    required String groupId,
    required String studentId,
    required String lessonDate,
    required AttendanceStatus status,
  }) async {
    markCalls++;
    await gate?.future;
    _maybeFail();
    marks[studentId] = status;
    return _rec(studentId);
  }

  @override
  Future<List<Attendance>> markMany({
    required String groupId,
    required String lessonDate,
    required Map<String, AttendanceStatus> marks,
  }) async {
    _maybeFail();
    this.marks.addAll(marks);
    return [for (final sid in marks.keys) _rec(sid)];
  }

  @override
  Future<void> delete(String groupId, String id) async {
    _maybeFail();
    marks.remove(id.substring(2));
  }

  @override
  Future<List<Attendance>> listStudent(String groupId, String studentId) async => [];
}

Future<RosterNotifier> _notifier(_FakeRepo repo) async {
  final n = RosterNotifier(repo, 'g');
  await Future<void>.delayed(Duration.zero); // первая загрузка
  return n;
}

/// Текущее состояние через подписку (debugState устарел, state защищён).
RosterState _st(RosterNotifier n) {
  late RosterState s;
  n.addListener((v) => s = v)();
  return s;
}

AttendanceStatus? _status(RosterNotifier n, String sid) =>
    _st(n).entries.firstWhere((e) => e.studentId == sid).status;

void main() {
  test('загружает всю группу, неотмеченные — без статуса', () async {
    final n = await _notifier(_FakeRepo());
    expect(_st(n).entries.map((e) => e.studentName), ['Aziza', 'Bekzod']);
    expect(_st(n).marked, 0);
    expect(_st(n).hasUnmarked, isTrue);
  });

  test('тап ставит статус, повторный тап по нему же снимает', () async {
    final repo = _FakeRepo();
    final n = await _notifier(repo);

    await n.toggle('s2', AttendanceStatus.absent);
    expect(_status(n, 's2'), AttendanceStatus.absent);
    expect(repo.marks['s2'], AttendanceStatus.absent);

    await n.toggle('s2', AttendanceStatus.absent);
    expect(_status(n, 's2'), isNull);
    expect(repo.marks.containsKey('s2'), isFalse);
  });

  test('тап по другому статусу меняет отметку', () async {
    final n = await _notifier(_FakeRepo());
    await n.toggle('s1', AttendanceStatus.present);
    await n.toggle('s1', AttendanceStatus.late);
    expect(_status(n, 's1'), AttendanceStatus.late);
  });

  test('«Остальные были» отмечает только неотмеченных', () async {
    final repo = _FakeRepo();
    final n = await _notifier(repo);
    await n.toggle('s1', AttendanceStatus.absent);

    await n.markRestPresent();
    expect(_status(n, 's1'), AttendanceStatus.absent, reason: 'чужую отметку не трогаем');
    expect(_status(n, 's2'), AttendanceStatus.present);
    expect(_st(n).hasUnmarked, isFalse);
  });

  test('без сети отметка откатывается и показывается ошибка', () async {
    final repo = _FakeRepo()..fail = true;
    final n = await _notifier(repo);

    await n.toggle('s1', AttendanceStatus.present);
    expect(_status(n, 's1'), isNull);
    expect(_st(n).saveError, isNotNull);

    await n.markRestPresent();
    expect(_st(n).marked, 0);
  });

  test('повторный тап, пока отметка сохраняется, игнорируется', () async {
    final repo = _FakeRepo()..gate = Completer<void>();
    final n = await _notifier(repo);

    final first = n.toggle('s1', AttendanceStatus.present);
    await n.toggle('s1', AttendanceStatus.absent); // ещё сохраняется
    repo.gate!.complete();
    await first;

    expect(repo.markCalls, 1);
    expect(_status(n, 's1'), AttendanceStatus.present);
  });

  test('будущие дни не открываются', () async {
    final n = await _notifier(_FakeRepo());
    final today = _st(n).date;
    n.shiftDay(1);
    expect(_st(n).date, today);
    n.shiftDay(-1);
    expect(_st(n).date, today.subtract(const Duration(days: 1)));
  });
}
