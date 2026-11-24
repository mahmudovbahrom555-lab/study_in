// Кнопка «Войти через Google»: на вебе её обязана рисовать сама Google (GIS),
// на Android — обычная кнопка, запускающая GoogleAuth.signIn.
export 'google_button_mobile.dart'
    if (dart.library.js_interop) 'google_button_web.dart';
