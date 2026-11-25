import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import '../../../core/localization/locale_provider.dart';

/// Черновик комментария к работе: хранится на устройстве, пока работа не
/// сдана, — чтобы текст не пропадал при выходе из задания.
abstract interface class HomeworkDraftStore {
  String? read(String key);
  Future<void> write(String key, String text);
  Future<void> clear(String key);
}

class HiveHomeworkDraftStore implements HomeworkDraftStore {
  HiveHomeworkDraftStore(this._box);

  final Box<dynamic> _box;

  @override
  String? read(String key) => _box.get('hw_draft:$key') as String?;

  @override
  Future<void> write(String key, String text) =>
      _box.put('hw_draft:$key', text);

  @override
  Future<void> clear(String key) => _box.delete('hw_draft:$key');
}

final homeworkDraftStoreProvider = Provider<HomeworkDraftStore>(
  (ref) => HiveHomeworkDraftStore(Hive.box<dynamic>(settingsBoxName)),
);
