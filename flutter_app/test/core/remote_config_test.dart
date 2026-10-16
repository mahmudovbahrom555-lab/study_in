import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/core/remote_config/remote_config.dart';

void main() {
  test('включены только фичи со значением true, неизвестные ключи не мешают', () {
    final c = RemoteConfig.fromJson({
      'min_version': '1.0.0',
      'features': {'quizzes': true, 'chat': false, 'future_feature': true},
    });
    expect(c.isEnabled(Feature.quizzes), isTrue);
    expect(c.isEnabled('chat'), isFalse);
    expect(c.isEnabled('geo'), isFalse);
  });

  test('ответ без features — всё выключено', () {
    final c = RemoteConfig.fromJson({'min_version': '0.0.0'});
    expect(c.features, isEmpty);
  });

  test('кэш читается обратно без потерь', () {
    const c = RemoteConfig(minVersion: '1.2.0', features: {'quizzes'});
    final back = RemoteConfig.fromJson(c.toJson());
    expect(back.minVersion, '1.2.0');
    expect(back.features, {'quizzes'});
  });

  test('обновление нужно только если версия ниже минимальной', () {
    const c = RemoteConfig(minVersion: '1.2.0');
    expect(c.requiresUpdate('1.1.9'), isTrue);
    expect(c.requiresUpdate('1.2.0'), isFalse);
    expect(c.requiresUpdate('1.10.0'), isFalse, reason: 'сравнение по числам, не по строке');
    expect(const RemoteConfig().requiresUpdate('0.0.1'), isFalse, reason: 'нет конфигурации');
  });

  test('compareVersions не учитывает build и pre-release суффиксы', () {
    expect(compareVersions('0.1.0+1', '0.1.0'), 0);
    expect(compareVersions('2.0.0-beta', '1.9.9'), greaterThan(0));
    expect(compareVersions('1.2', '1.2.0'), 0);
  });
}
