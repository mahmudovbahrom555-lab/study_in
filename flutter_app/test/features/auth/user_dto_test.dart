import 'package:flutter_test/flutter_test.dart';
import 'package:mamu_learn/features/auth/data/models/user_dto.dart';

void main() {
  test('вошедший через Google: телефона нет, есть email', () {
    final user = UserDto.fromJson({
      'id': 'u1',
      'phone': '',
      'email': 'a@gmail.com',
      'name': 'Ali',
      'language': 'uz',
      'is_active': true,
    }).toEntity();

    expect(user.phone, isEmpty);
    expect(user.email, 'a@gmail.com');
    expect(user.copyWith(name: 'Vali').email, 'a@gmail.com');
  });

  test('вошедший по телефону: email в ответе нет', () {
    final user = UserDto.fromJson({
      'id': 'u2',
      'phone': '+998901234567',
      'name': 'Ali',
    }).toEntity();

    expect(user.phone, '+998901234567');
    expect(user.email, isEmpty);
  });
}
