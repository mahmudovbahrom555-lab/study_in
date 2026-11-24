import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import '../../../core/localization/locale_provider.dart';

/// Где хранится год рождения с возрастного экрана.
abstract interface class BirthYearStore {
  int? read();
  Future<void> write(int year);
}

class HiveBirthYearStore implements BirthYearStore {
  HiveBirthYearStore(this._box);

  static const _key = 'birth_year';
  final Box<dynamic> _box;

  @override
  int? read() => _box.get(_key) as int?;

  @override
  Future<void> write(int year) => _box.put(_key, year);
}

/// Год рождения с нейтрального возрастного экрана (политика Google Play для
/// семей). Хранится только на устройстве, на сервер не отправляется.
class BirthYearController extends StateNotifier<int?> {
  BirthYearController(this._store) : super(_store.read());

  final BirthYearStore _store;

  Future<void> select(int year) async {
    state = year;
    await _store.write(year);
  }
}

final birthYearProvider = StateNotifierProvider<BirthYearController, int?>(
  (ref) => BirthYearController(
    HiveBirthYearStore(Hive.box<dynamic>(settingsBoxName)),
  ),
);

/// Детям младше 13 вход через Google не показываем. Месяц не спрашиваем,
/// поэтому родившиеся в год, когда им исполняется 13, считаются детьми.
bool isChildBirthYear(int year, {DateTime? now}) =>
    (now ?? DateTime.now()).year - year <= 13;
