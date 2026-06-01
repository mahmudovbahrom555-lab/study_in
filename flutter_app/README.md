# Flutter App (RepetApp)

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
flutter run --release                    # Release режим
flutter analyze                          # Линтер
flutter test                             # Тесты
dart format lib/                         # Форматирование
flutter pub run build_runner build       # Кодогенерация
```

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
