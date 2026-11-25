import 'dart:js_interop';

import 'package:web/web.dart' as web;

const _key = 'mamu.session.user';

/// Какой аккаунт вошёл в этой вкладке — для остальных вкладок.
void publishSessionUser(String? userId) {
  try {
    final storage = web.window.localStorage;
    if (userId == null) {
      storage.removeItem(_key);
    } else {
      storage.setItem(_key, userId);
    }
  } catch (_) {
    // localStorage недоступен (запрет в браузере) — синхронизации не будет.
  }
}

/// Событие storage браузер присылает только в другие вкладки сайта.
/// key == null — хранилище очистили целиком.
void onSessionUserChangedElsewhere(void Function(String? userId) callback) {
  void onStorage(web.StorageEvent e) {
    if (e.key == _key || e.key == null) callback(e.newValue);
  }

  web.window.addEventListener('storage', onStorage.toJS);
}

void reloadApp() => web.window.location.reload();
