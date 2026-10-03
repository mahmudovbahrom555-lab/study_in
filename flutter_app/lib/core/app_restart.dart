import 'package:flutter/widgets.dart';

/// Позволяет перезапустить всё приложение: пересоздаёт поддерево вместе с
/// ProviderScope, так что все провайдеры (группы, оценки, уведомления…)
/// начинают с чистого листа.
///
/// Нужно при выходе из аккаунта: иначе следующий пользователь на этом же
/// устройстве увидит закэшированные данные предыдущего.
class AppRestartScope extends StatefulWidget {
  const AppRestartScope({super.key, required this.child});

  final Widget child;

  static void restart(BuildContext context) =>
      context.findAncestorStateOfType<_AppRestartScopeState>()?._restart();

  @override
  State<AppRestartScope> createState() => _AppRestartScopeState();
}

class _AppRestartScopeState extends State<AppRestartScope> {
  Key _key = UniqueKey();

  void _restart() => setState(() => _key = UniqueKey());

  @override
  Widget build(BuildContext context) =>
      KeyedSubtree(key: _key, child: widget.child);
}
