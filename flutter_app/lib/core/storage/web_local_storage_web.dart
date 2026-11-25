import 'package:web/web.dart' as web;

String? webRead(String key) => web.window.localStorage.getItem(key);

void webWrite(String key, String value) =>
    web.window.localStorage.setItem(key, value);

void webDelete(String key) => web.window.localStorage.removeItem(key);
