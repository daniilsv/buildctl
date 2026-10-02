# Build Assistant

Система для отслеживания и управления сборками проектов с интеграцией AI для анализа коммитов и автоматическими уведомлениями.

## Состав проекта

Проект состоит из трех основных компонентов:

- **Backend** (`back/`) - Go сервер с REST API для управления проектами, ветками, сборками и артефактами
- **Frontend** (`front/`) - React веб-интерфейс для просмотра и управления сборками

## Функциональность

- Отслеживание сборок по проектам и веткам
- Хранение логов сборок в БД
- Управление артефактами сборок:
  - Загрузка файловых артефактов в S3 с публичными URL
  - Регистрация образов контейнеров с тегами и дайджестами
  - Просмотр и скачивание артефактов в веб-интерфейсе
  - Удаление артефактов (отдельных или всех для сборки)
- AI-анализ коммитов с генерацией описаний на русском языке
- Уведомления в Telegram о статусе сборок с информацией об артефактах
- OIDC аутентификация для веб-интерфейса
- REST API для интеграции с внешними системами

## Переменные окружения для запуска

### Backend

```bash
DATABASE_URL=postgres://user:password@host:5432/build_assistant?sslmode=disable
PORT=8080

# OIDC (для аутентификации)
OIDC_ISSUER=https://your-oidc-provider.com
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=https://your-frontend.com/auth/callback

# OpenAI/OpenRouter (для AI анализа)
OPENAI_API_URL=https://openrouter.ai/api/v1
OPENAI_API_KEY=your-api-key
OPENAI_MODEL=qwen/qwen3-235b-a22b-2507

# S3 (для хранения артефактов)
S3_ENDPOINT=https://s3.example.com
S3_REGION=us-east-1
S3_PATH_STYLE=true
S3_BUCKET=build-assistant-artifacts
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_PUBLIC_PREFIX=https://cdn.example.com

# Telegram (для уведомлений)
TELEGRAM_BOT_TOKEN=your-bot-token

# Worker pool
WORKER_POOL_SIZE=3
```

## Запуск через Docker Compose

### Требования

- Docker и Docker Compose
- PostgreSQL база данных (должна быть запущена отдельно, не включена в compose)

### Запуск

1. Скопируйте `.infra/compose.yml` и создайте `.env` файл с переменными окружения:

```bash
cd .infra
cp compose.yml docker-compose.yml
```

2. Создайте `.env` файл с необходимыми переменными (см. раздел выше)

3. Запустите сервисы:

```bash
docker-compose up -d
```

Сервисы будут доступны:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080 (внутренний порт, не экспортируется наружу)

### Структура compose

- `backend` - Go сервер API
- `frontend` - React приложение на nginx

**Примечание:** База данных PostgreSQL должна быть запущена отдельно и доступна по `DATABASE_URL`. Порт бэкенда не экспортируется наружу в production compose.

## Разработка

Для локальной разработки используйте `compose.local.yml`, который включает PostgreSQL и экспортирует порты для отладки.

## API Endpoints для артефактов

### Получение presigned URL для загрузки

```http
POST /api/v1/artifacts/presign
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_name": "my-project",
  "branch_name": "main",
  "commit_hash": "abc123",
  "filename": "app.tar.gz"
}
```

Ответ:
```json
{
  "upload_url": "https://s3.example.com/...",
  "s3_key": "builds/abc123/app.tar.gz"
}
```

### Подтверждение загрузки файлового артефакта

```http
POST /api/v1/artifacts/confirm
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_name": "my-project",
  "branch_name": "main",
  "commit_hash": "abc123",
  "s3_key": "builds/abc123/app.tar.gz",
  "filename": "app.tar.gz",
  "size_bytes": 1024000,
  "content_type": "application/gzip",
  "log_id": "optional-log-uuid"
}
```

### Регистрация образа контейнера

```http
POST /api/v1/artifacts/container
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_name": "my-project",
  "branch_name": "main",
  "commit_hash": "abc123",
  "image_name": "registry.example.com/myapp:v1.0.0",
  "image_tag": "v1.0.0",
  "image_digest": "sha256:abc123...",
  "filename": "optional-image.tar",
  "s3_key": "optional-s3-key",
  "size_bytes": 50000000,
  "log_id": "optional-log-uuid"
}
```

### Получение артефактов сборки

```http
GET /api/v1/builds/{build_id}/artifacts
Authorization: Bearer <token>
```

Ответ:
```json
[
  {
    "id": "uuid",
    "build_id": "uuid",
    "artifact_type": "file",
    "filename": "app.tar.gz",
    "size_bytes": 1024000,
    "content_type": "application/gzip",
    "public_url": "https://cdn.example.com/builds/abc123/app.tar.gz",
    "created_at": "2025-01-01T12:00:00Z"
  },
  {
    "id": "uuid",
    "build_id": "uuid",
    "artifact_type": "container_image",
    "image_name": "registry.example.com/myapp:v1.0.0",
    "image_tag": "v1.0.0",
    "image_digest": "sha256:abc123...",
    "created_at": "2025-01-01T12:05:00Z"
  }
]
```

### Удаление артефакта

```http
DELETE /api/v1/artifacts/{artifact_id}
Authorization: Bearer <token>
```

### Удаление всех артефактов сборки

```http
DELETE /api/v1/builds/{build_id}/artifacts
Authorization: Bearer <token>
```


## GitHub Actions: сборка Docker

Workflow `.github/workflows/docker-build.yml` собирает единый образ из
`.infra/Dockerfile` на каждый push и pull request. Образ не публикуется,
приложение не запускается; секреты и deployment не требуются.

Локальная проверка:

```bash
docker build --platform linux/amd64 -f .infra/Dockerfile -t buildctl:local .
```

Локальные credentials задавайте через переменные окружения. Конфигурация
`.vscode/launch.json` читает их через `${env:VARIABLE}`. Не коммитьте секреты.
История этой копии очищена от прежнего файла с credentials; исходный
репозиторий Тауруса сохранён отдельно, поэтому SHA исторических коммитов отличаются.
