// Веб: токены — в обычном localStorage. flutter_secure_storage шифрует их
// ключом, который лежит там же, в браузере, — защиты это не даёт, а при
// записях из нескольких вкладок ключи расходятся, значение не расшифровывается
// и вход стирается («выкидывает из аккаунта» после обновления страницы).
export 'web_local_storage_stub.dart'
    if (dart.library.js_interop) 'web_local_storage_web.dart';
