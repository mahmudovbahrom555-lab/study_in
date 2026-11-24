# Политика конфиденциальности MaMu Learn

Опубликованный текст — единственный источник: [`deploy/site/privacy.html`](../../deploy/site/privacy.html),
адрес https://mamu.education/privacy. Удаление аккаунта (ссылка для Google Play):
[`deploy/site/delete-account.html`](../../deploy/site/delete-account.html), https://mamu.education/delete-account.

Решения владельцев (2026-11-24): публикуется как есть, без проверки юристом; операторы —
Shukurov Murod и Sodiqjonova Asilya; связь — mahmudovbahrom555@gmail.com.

При изменении текста:
1. Поменять дату и версию в начале страницы.
2. Поменять `AppConfig.privacyVersion` во Flutter на ту же версию — сервер хранит в журнале
   согласий, с какой версией согласился пользователь.
3. Перед включением SMS (Eskiz), push (Firebase), отчётов об ошибках (Sentry), AI (OpenAI)
   или проксирования через Cloudflare — дописать сервис в таблицу раздела 4.

Открытые вопросы, о которых владельцы знают: ст. 27-1 ЗРУ-547 (хранение данных граждан
Узбекистана в Узбекистане — сервер сейчас в США), возраст согласия ребёнка, версии на
узбекском и английском.
