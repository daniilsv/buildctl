# Node-RED Flow: Уведомление о билде

## Назначение
Автоматическое уведомление о билде в Telegram с кратким описанием коммитов на русском языке

## Поток данных

### 1. HTTP Endpoint
**Точка входа:** `POST /notify_build`

Принимает JSON с параметрами:
- `repo` - репозиторий (по умолчанию: "sd/sd-back")
- `branch` - ветка (по умолчанию: "develop")
- `start_hash` - начальный SHA коммита
- `end_hash` - конечный SHA коммита

### 2. Подготовка запроса к Git API
**Функция:** set git url, save initial payload

Действия:
- Устанавливает значения по умолчанию для repo и branch
- Сохраняет исходный payload в `msg.initial_payload`
- Формирует URL: `/repos/{repo}/commits?sha={branch}&limit=100`

### 3. Запрос коммитов из Git
**API:** axios-request к git.int.sktaurus.ru

Параметры:
- Method: GET
- Endpoint: `/repos/sd/sd-back/commits?sha=develop&limit=100`
- Timeout: 30 секунд
- Response type: JSON

### 4. Фильтрация коммитов
**Функция:** filter commits

Алгоритм:
- Находит коммит с `start_hash`
- Собирает все коммиты от start_hash до end_hash
- Для каждого коммита сохраняет: SHA, сообщение, автор, дата

### 5. Подготовка промпта для AI
**Функция:** prepare summary and translate

Действия:
- Формирует запрос к OpenAI API (модель: qwen/qwen3-235b-a22b-2507)
- Промпт: "Суммировать коммиты в 3 предложения и перевести на русский"
- Max tokens: 8096
- Передает все сообщения коммитов
{
    "model": "qwen/qwen3-235b-a22b-2507",
    "max_tokens": 8096,
    "messages": [
        {
            "role": "user",
            "content": [
                {
                    "type": "text",
                    "text": "You are mighty git commit summarizer and translator. Next i provide you a bunch of commit messages. Your role is to summarize them all in one with no more than 3 sentences at all and translate to Russian. In response give only translated summarized message that i can send to report. Omit word Commit, any markdown or styling!\n\n"+
                    "Commits:\n"+msg.commits.map(com=>com.message).join("\n")
                }
            ]
        }
    ]
}

### 6. Обработка AI
**OpenAI API:** chat completion

Модель обрабатывает сообщения и возвращает краткую суммаризацию на русском языке

### 7. Формирование сообщения для Telegram
**Функция:** fill tg message

Создает payload:
- `chatId`: -4859320492
- `type`: "message"
- `content`: результат от AI

### 8. Отправка в Telegram
**Telegram sender:** Taurus IT Assistant (@taurus_service_sd_bot)

Отправляет финальное сообщение в указанный чат

## Конфигурация

### API Endpoints
- **Git API**: https://git.int.sktaurus.ru/api/v1/
- **OpenAI**: https://openrouter.ai/api/v1

### Telegram Bot
- **Имя**: Taurus IT Assistant
- **Username**: @taurus_service_sd_bot
- **Chat ID**: -4859320492
- **Update mode**: polling (интервал 300ms)

### Модули
- node-red-contrib-axios: 1.6.0
- @inductiv/node-red-openai-api: 1.103.0-patch.1
- node-red-contrib-telegrambot: 16.3.2
