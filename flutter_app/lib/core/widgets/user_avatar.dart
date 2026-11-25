import 'package:flutter/material.dart';

/// Аватары-животные: ключ с сервера → эмодзи. Тот же набор, что
/// domain.AvatarPresets на сервере.
const avatarPresets = {
  'fox': '🦊',
  'panda': '🐼',
  'koala': '🐨',
  'lion': '🦁',
  'tiger': '🐯',
  'frog': '🐸',
  'monkey': '🐵',
  'penguin': '🐧',
  'bear': '🐻',
  'rabbit': '🐰',
  'cat': '🐱',
  'dog': '🐶',
};

/// Аватар пользователя: животное, фото или первая буква имени.
class UserAvatar extends StatelessWidget {
  const UserAvatar({
    super.key,
    required this.name,
    this.avatarUrl,
    this.avatarPreset,
    this.radius = 20,
  });

  final String name;
  final String? avatarUrl;
  final String? avatarPreset;
  final double radius;

  @override
  Widget build(BuildContext context) {
    final emoji = avatarPresets[avatarPreset];
    if (emoji != null) {
      return CircleAvatar(
        radius: radius,
        backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
        child: Text(emoji, style: TextStyle(fontSize: radius)),
      );
    }
    final url = avatarUrl;
    if (url != null) {
      return CircleAvatar(radius: radius, backgroundImage: NetworkImage(url));
    }
    final initial = name.isEmpty ? '?' : name.substring(0, 1).toUpperCase();
    return CircleAvatar(radius: radius, child: Text(initial));
  }
}
