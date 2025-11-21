# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы с управлением командами и участниками.
[![CI](https://github.com/darik1201/AvitoIntern_Task/actions/workflows/ci.yml/badge.svg)](https://github.com/darik1201/AvitoIntern_Task/actions)
## Быстрый старт

```bash
docker-compose up --build
```

Сервис доступен на `http://localhost:8080`. Миграции применяются автоматически.

## Описание

Сервис автоматически назначает до 2 ревьюеров на PR из команды автора, позволяет выполнять переназначение ревьюверов и получать список PR'ов, назначенных конкретному пользователю. После merge PR изменение состава ревьюверов запрещено.

## Команды

```bash
make build           # Собрать проект
make run             # Запустить локально
make test            # Запустить тесты
make lint            # Запустить линтер
make docker-up       # Запустить через Docker Compose
make docker-down     # Остановить Docker Compose
make swagger         # Сгенерировать Swagger документацию
make load-test       # Запустить нагрузочное тестирование
```

## API Эндпоинты

### Teams
- `POST /team/add` - Создать команду с участниками
- `GET /team/get?team_name=<name>` - Получить команду
- `POST /team/bulkDeactivate` - Массовая деактивация всех участников команды

### Users
- `POST /users/setIsActive` - Установить флаг активности пользователя
- `GET /users/getReview?user_id=<id>` - Получить PR'ы, где пользователь назначен ревьювером

### Pull Requests
- `POST /pullRequest/create` - Создать PR и автоматически назначить до 2 ревьюеров
- `POST /pullRequest/merge` - Пометить PR как MERGED (идемпотентная операция)
- `POST /pullRequest/reassign` - Переназначить ревьювера на другого из его команды

### Дополнительные
- `GET /stats` - Статистика назначений по пользователям и PR'ам
- `GET /health` - Health check с проверкой БД
- `GET /metrics` - Prometheus метрики
- `GET /swagger` - Swagger UI документация

## Фичи

### Structured Logging
- JSON логирование всех HTTP запросов (zerolog)
- Request ID для трейсинга через X-Request-ID header
- Детальная информация: method, path, status, latency, IP, user-agent

### Prometheus Metrics
- Метрики HTTP запросов (количество, продолжительность)
- Доступны на `/metrics`
- Готовы для интеграции с Prometheus/Grafana

### Graceful Shutdown
- Корректное завершение работы при SIGINT/SIGTERM
- Ожидание завершения активных запросов (таймаут 10 сек)
- Закрытие соединений с БД

### Health Check
- Проверка доступности PostgreSQL с таймаутом 2 сек
- Возвращает статус приложения и БД
- Готов для Kubernetes/Docker health checks

### Middleware Stack
- **Recovery** - обработка паник без падения сервера
- **RequestID** - уникальный ID для каждого запроса
- **Logger** - структурированное логирование
- **PrometheusMetrics** - сбор метрик

### Swagger UI
- Интерактивная документация API
- Доступна на `/swagger`
- Генерация из OpenAPI спецификации

## Тестирование

### Интеграционные тесты

12 интеграционных тестов покрывают все основные эндпоинты:

```bash
make test
```

Покрытие:
- Health check, Teams, Users, Pull Requests
- Проверка ошибок (404, 409, 400)
- Бизнес-логика (автоназначение, идемпотентность, запрет изменений после MERGED)

### Нагрузочное тестирование

Используется K6 для проверки производительности:

```bash
make load-test
```

Тест проверяет:
- Время ответа < 300ms
- Частота ошибок < 1%
- Градуальная нагрузка: 0 → 10 → 50 → 100 → 50 → 0 пользователей

### Ручное тестирование

Скрипт для автоматического тестирования всех эндпоинтов:


## Мониторинг

### Метрики
Готовы для интеграции с:
- **Prometheus** - сбор метрик
- **Grafana** - визуализация
- **AlertManager** - алерты

### Логи
Логи в JSON формате, готовы для:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Loki** + Grafana
- **CloudWatch**
- Любых других систем централизованного логирования

## CI/CD

### GitHub Actions

- **CI Pipeline** (`.github/workflows/ci.yml`):
  - Запуск тестов с PostgreSQL
  - Линтинг кода (golangci-lint)
  - Сборка проекта
  - Docker build проверка

- **Migration Check** (`.github/workflows/migrate.yml`):
  - Проверка миграций при изменении файлов в `migrations/`

- **Release** (`.github/workflows/release.yml`):
  - Создание релизов для всех платформ (Linux, macOS, Windows)
  - Автоматически при создании тега `v*.*.*`

- **Dependabot** (`.github/dependabot.yml`):
  - Автоматическое обновление зависимостей Go, Docker и GitHub Actions

### Линтер

Настроен `golangci-lint` с конфигурацией в `.golangci.yml`:
- errcheck, gosimple, govet, staticcheck
- goconst, gocritic, gocyclo
- goimports, misspell, revive

## Структура проекта

```
.
├── cmd/server/              # Точка входа приложения
├── internal/
│   ├── database/           # Подключение к БД
│   ├── handler/            # HTTP handlers
│   ├── middleware/         # Middleware (logging, metrics, recovery)
│   ├── models/             # Модели данных
│   ├── repository/         # Слой работы с БД
│   ├── router/             # Роутинг
│   ├── service/            # Бизнес-логика
│   └── test/               # Интеграционные тесты
├── migrations/             # SQL миграции
├── k6/                     # K6 скрипты для нагрузочного тестирования
├── .github/
│   ├── workflows/          # GitHub Actions workflows
│   └── dependabot.yml      # Автообновление зависимостей
├── openapi.yml            # OpenAPI спецификация
├── docker-compose.yml     # Docker Compose конфигурация
├── Dockerfile             # Docker образ приложения
├── Makefile              # Команды для сборки и запуска
└── README.md             # Документация
```

## Архитектура

- **Handler** → **Service** → **Repository**
- Чистая архитектура с разделением слоёв
- Dependency injection через конструкторы
- Обработка ошибок согласно OpenAPI спецификации

## Переменные окружения

```bash
DB_HOST=postgres          # Хост БД
DB_PORT=5435              # Порт БД
DB_USER=postgres          # Пользователь БД
DB_PASSWORD=postgres      # Пароль БД
DB_NAME=pr_reviewer       # Имя БД
DB_SSLMODE=disable        # SSL режим
PORT=8080                 # Порт сервера
```

## Примеры использования

### Создать команду и PR

```bash
# Создать команду
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'

# Создать PR (автоматически назначит 2 ревьюеров из команды)
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "pull_request_name": "Add feature",
    "author_id": "u1"
  }'
```

### Получить метрики и статистику

```bash
# Prometheus метрики
curl http://localhost:8080/metrics

# Статистика
curl http://localhost:8080/stats | jq .
```

### Swagger UI

Откройте в браузере: `http://localhost:8080/swagger`


## Технологии

- **Язык**: Go 1.23
- **База данных**: PostgreSQL 16
- **Фреймворк**: Gin
- **Логирование**: zerolog
- **Метрики**: Prometheus
- **Миграции**: golang-migrate
- **Линтер**: golangci-lint

## Лицензия

MIT
