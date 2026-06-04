# Link Storage Service

Тестовое задание на Go.

Сервис позволяет:

- создавать короткие ссылки
- получать оригинальный URL по short_code
- получать статистику
- удалять ссылки
- получать список ссылок с пагинацией

В качестве хранилища используется Redis.

## Запуск

```bash
docker compose up --build
```

После запуска сервис доступен по адресу:

http://localhost:8080

## Примеры запросов

Создание ссылки:

```bash
curl -X POST http://localhost:8080/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/test"}'
```

Получение ссылки:

```bash
curl http://localhost:8080/links/{short_code}
```

Получение статистики:

```bash
curl http://localhost:8080/links/{short_code}/stats
```

Получение списка ссылок:

```bash
curl http://localhost:8080/links?limit=10&offset=0
```

Удаление ссылки:

```bash
curl -X DELETE http://localhost:8080/links/{short_code}
```

## Конфигурация

Поддерживаются следующие переменные окружения:

- APP_PORT
- REDIS_ADDR
- REDIS_PASSWORD
- REDIS_DB

## Структура проекта

```text
cmd/app              - точка входа
internal/handler     - HTTP обработчики
internal/service     - бизнес-логика
internal/repository  - работа с Redis
internal/cache       - in-memory cache
internal/model       - модели данных
```

