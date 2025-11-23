# Build Assistant

Система для отслеживания и управления сборками проектов с интеграцией AI для анализа коммитов и автоматическими уведомлениями.

## Состав проекта

Проект состоит из трех основных компонентов:

- **Backend** (`back/`) - Go сервер с REST API для управления проектами, ветками, сборками и артефактами
- **Frontend** (`front/`) - React веб-интерфейс для просмотра и управления сборками
- **CLI** (`cli/`) - утилита `buildctl` для интеграции с CI/CD системами

## Функциональность

- Отслеживание сборок по проектам и веткам
- Хранение логов сборок и артефактов в S3
- AI-анализ коммитов с генерацией описаний на русском языке
- Уведомления в Telegram о статусе сборок
- OIDC аутентификация для веб-интерфейса
- REST API для интеграции с внешними системами

## Использование CLI в GitHub Actions

### Пример workflow

```yaml
name: Build and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Notify build started
        run: |
          buildctl event \
            --token="${{ secrets.BUILD_ASSISTANT_TOKEN }}" \
            --backend="${{ secrets.BUILD_ASSISTANT_BACKEND }}" \
            --project="my-project" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="started"
      
      - name: Run tests
        run: |
          # Ваши команды сборки/тестирования
          npm test
        continue-on-error: true

      - name: Notify build step
        run: |
          buildctl event \
            --token="${{ secrets.BUILD_ASSISTANT_TOKEN }}" \
            --backend="${{ secrets.BUILD_ASSISTANT_BACKEND }}" \
            --project="my-project" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="First service tested"

      - name: Run tests
        run: |
          # Ваши команды сборки/тестирования
          npm build
        continue-on-error: true

      - name: Upload artifact
        if: success()
        run: |
          buildctl artifact upload \
            --token="${{ secrets.BUILD_ASSISTANT_TOKEN }}" \
            --backend="${{ secrets.BUILD_ASSISTANT_BACKEND }}" \
            --project="my-project" \
            --branch="${{ github.ref_name }}" \
            --commit="${{ github.sha }}" \
            --file="./dist/app.tar.gz"
      
      - name: Notify build success
        if: success()
        run: |
          buildctl event \
            --token="${{ secrets.BUILD_ASSISTANT_TOKEN }}" \
            --backend="${{ secrets.BUILD_ASSISTANT_BACKEND }}" \
            --project="my-project" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="success"
      
      - name: Notify build failed
        if: failure()
        run: |
          buildctl event \
            --token="${{ secrets.BUILD_ASSISTANT_TOKEN }}" \
            --backend="${{ secrets.BUILD_ASSISTANT_BACKEND }}" \
            --project="my-project" \
            --commit="${{ github.sha }}" \
            --branch="${{ github.ref_name }}" \
            --status="failed"
      
```

### Переменные окружения для GitHub Secrets

Необходимо добавить в Settings → Secrets and variables → Actions:

- `BUILD_ASSISTANT_TOKEN` - токен доступа к API (создается в веб-интерфейсе)
- `BUILD_ASSISTANT_BACKEND` - URL бэкенда (например, `https://build-assistant.example.com`)

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

## CLI команды

### Отправка события сборки

```bash
buildctl event \
  --token="your-token" \
  --backend="https://backend.example.com" \
  --project="project-name" \
  --commit="abc123" \
  --branch="main" \
  --status="started|in_progress|success|failed|cancelled" \
  --log="Optional log message"
```

### Загрузка артефакта

```bash
buildctl artifact upload \
  --token="your-token" \
  --backend="https://backend.example.com" \
  --project="project-name" \
  --branch="main" \
  --commit="abc123" \
  --file="./path/to/artifact.tar.gz"
```

