# Деплой на Contabo (прод/тест на телефоне)

## 1. Сервер
1. contabo.com → **Cloud VPS** (VPS 10 или 20, Ubuntu 22.04 LTS). Оплата картой; если не
   проходит — попробовать PayPal.
2. Получить root SSH-доступ и публичный IP из письма/панели Contabo.

## 2. Домен (Cloudflare)
В DNS уже купленного домена добавить (проксирование — оранжевое облако — выключить на этом шаге):
- `A  @    <IP сервера>`
- `A  api  <IP сервера>`

## 3. Установка Docker на сервере
```bash
ssh root@<IP>
apt update && apt install -y docker.io docker-compose-plugin git
systemctl enable --now docker
```

## 4. Клонирование и настройка
```bash
git clone https://github.com/mahmudovbahrom555-lab/study_in.git
cd study_in
cp .env.prod.example .env.prod
nano .env.prod   # заполнить пароли, JWT-секреты (openssl rand -hex 32), домены
```

## 5. Сборка Flutter web (локально, перед пушем или прямо на сервере)
```bash
cd flutter_app
flutter build web --dart-define=API_BASE_URL=https://api.yourdomain.com
cd ..
```
Caddy отдаёт `flutter_app/build/web` как статику — папка должна существовать на сервере
до старта контейнеров.

## 6. Запуск
```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
docker compose -f docker-compose.prod.yml ps   # все healthy?
```
Caddy сам получит и обновит Let's Encrypt сертификаты для `API_DOMAIN`/`WEB_DOMAIN` —
для этого порты 80/443 должны быть открыты и домены уже указывать на сервер (шаг 2).

## 7. Проверка
- `https://api.yourdomain.com/api/v1/health` → 200 OK
- `https://yourdomain.com` в браузере на телефоне → открывается web-версия

Теперь можно включить проксирование Cloudflare (оранжевое облако) обратно.

## 8. Android APK
Локально:
```bash
cd flutter_app
flutter build apk --release --dart-define=API_BASE_URL=https://api.yourdomain.com
```
Файл: `flutter_app/build/app/outputs/flutter-apk/app-release.apk`.
Залить его на сервер в `flutter_app/build/web/downloads/app-release.apk` (или через
Cloudflare Pages/R2 отдельно) — Caddy отдаст прямую ссылку для скачивания на телефоне.

## 9. Обновление после изменений в коде
```bash
cd study_in
git pull
flutter build web --dart-define=API_BASE_URL=https://api.yourdomain.com   # если менялся flutter_app
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
```

## Примечания
- Автодеплоя нет — всё руками через SSH (так и задумано, см. `AI_CONTEXT.md`).
- `.env.prod` никогда не коммитить — только `.env.prod.example`.
- MinIO хранит файлы пользователей на сервере — для реального продакшена со временем
  стоит рассмотреть внешний S3 (Cloudflare R2) ради бэкапов, но для теста на телефоне
  текущей связки достаточно.
