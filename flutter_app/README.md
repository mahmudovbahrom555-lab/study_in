# Flutter App (MaMu Learn)

## Структура

```
lib/
├── main.dart                 # Точка входа
├── app.dart                  # Корневой виджет
├── core/                     # Общее ядро
│   ├── config/               # AppConfig
│   ├── theme/                # Темы, цвета
│   ├── router/               # Навигация (Этап 1+)
│   ├── network/              # Dio, интерсепторы
│   ├── storage/              # Hive, secure storage
│   ├── localization/         # ARB файлы
│   ├── error/                # Failure классы
│   └── utils/                # Хелперы
├── features/                 # Изолированные фичи
│   ├── auth/
│   │   ├── data/             # API, репозитории
│   │   ├── domain/           # Сущности, use cases
│   │   └── presentation/     # UI, провайдеры
│   ├── groups/               # (Этап 2)
│   ├── assignments/          # (Этап 4)
│   └── ...
├── shared/                   # Переиспользуемые виджеты
│   └── widgets/
└── generated/                # Авто-генерируемый код (l10n, freezed)
```

## Быстрый старт

```bash
flutter pub get
flutter run --dart-define=API_URL=http://localhost:8080/api/v1
```

Для iOS симулятора используйте `http://localhost:8080/api/v1`.
Для Android эмулятора — `http://10.0.2.2:8080/api/v1`.

## Команды

```bash
flutter pub get                          # Установить зависимости
flutter run                              # Запустить (debug)
flutter analyze                          # Линтер
flutter test                             # Тесты
dart format lib/                         # Форматирование
flutter pub run build_runner build       # Кодогенерация
```

## Android App Bundle для Google Play

Используйте Flutter 3.38.9, JDK 17 и Android SDK Platform 36.
В `android/key.properties` укажите `storeFile` (абсолютный путь или путь
относительно `android/`), `storePassword`, `keyAlias` и `keyPassword`.
Используйте ключ загрузки для приложения `uz.mamulearn.app` в Play Console.
Файл настроек и хранилище ключей должны оставаться вне Git.

```bash
flutter pub get --enforce-lockfile
flutter build appbundle --release \
  --dart-define=ENV=production \
  --dart-define=API_URL=https://api.mamu.education/api/v1
```

Пакет для загрузки будет в `build/app/outputs/bundle/release/app-release.aab`.
Перед повторной загрузкой увеличьте номер сборки после `+` в `pubspec.yaml`.
План проверок перед публикацией описан в [плане релиза](../docs/PLAY-RELEASE.md).

## Принципы

- **Feature-First**: каждая фича — изолированный модуль с data/domain/presentation
- **Riverpod**: для управления состоянием и DI
- **Чистая архитектура внутри фичи**: data → domain → presentation
- **Локализация**: 4 языка (uz, uz-Cyrl, ru, en) через ARB файлы

## Добавление новой фичи

1. Создать `lib/features/<name>/data/`, `domain/`, `presentation/`
2. В `domain/entities/` — модели сущностей
3. В `domain/repositories/` — интерфейс репозитория
4. В `data/` — реализация репозитория и API
5. В `presentation/providers/` — Riverpod провайдеры
6. В `presentation/pages/` — экраны
7. Зарегистрировать роуты в `core/router/`
