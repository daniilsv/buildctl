У нас при сборке мобильного приложения, веб-приложения и бэкенда запускаются после.в CI/CD.Небольшие обработки для рассылки уведомлений о том, что собрана новая сборка.Плюс посылаются вебхуки для обновления контейнеров.и может что то еще тоже какие-то хуки.я хочу добавить в pipeline и ci/cd гите.но на деле неважно куда именно.свой обработчик, который будет сласть некие сообщения в систему, которую я хочу сделать.Чтобы в ней уже по.хэшу коммита.определялось.Что изменилось?Плюс по окончании такой сборки рассылались как раз таки все уведомления, вебхуки, обновления.Контейнеров и прочее, проче.Мне нужно, чтобы ты вместе со мной поштурмил.может, задавал какие то вопросы.чтобы в итоге выйти на реальный план.подобной реализации, чтобы потом скормить это все ай агенту, чтобы он смог целиком практически.Сделать такую систему.Можно разделить все это на блоки.которые будут создаваться.По отдельности.Но в итоге это должно быть.как я вижу.Вернись.Front.и небольшая программа, которая сможет принимать на вход.аргументы.С токеном.доступа.адресом бека.идентификатором проекта.И хэш коммита.И статусом, может логом.чтобы можно было эту программу выполнять, как в начале САИ.пайплайн, так и в промежутке, так и на финале.

---

Отличная задача! Давайте вместе пройдем путь от анализа до конкретного плана. Я начну с уточняющих вопросов, чтобы закрыть неопределенности, а затем предложу пошаговую реализацию.

---

### 🔍 **Шаг 1: Уточняющие вопросы**
Чтобы спроектировать систему правильно, нужно понять детали:

1. **CI/CD-инфраструктура**  
   Какой инструмент вы используете? (GitLab CI, GitHub Actions, Jenkins, Argo CD и т.д.)  
   *→ Это повлияет на способ интеграции CLI-утилиты в пайплайн и работу с секретами.*

2. **Данные для анализа**  
   Помимо хеша коммита, какие данные нужны вашей системе? Например:  
   - Список измененных файлов?  
   - Diff коммита?  
   - Автор коммита?  
   - Ветка/тег?  
   *→ Это определит, как бэкенд будет взаимодействовать с Git API.*

3. **Безопасность**  
   Как аутентифицировать запросы к вашему бэкенду?  
   - Токен в заголовке `Authorization: Bearer <token>`

4. **Обработка ошибок**  
   Что делать, если:  
   - CLI-утилита не может отправить данные в бэкенд?  
   - Рассылка вебхуков завершилась ошибкой?  
   *→ Нужны ли повторные попытки? Локальное кэширование?*

5. **Асинхронность**  
   Должна рассылка уведомлений и обновление контейнеров происходить асинхронно (через очередь задач). NATS

6. **Логирование**  
   Куда писать логи CLI-утилиты?  
   - В stdout/stderr (для захвата CI)

---

### 🧱 **Шаг 2: Архитектурные блоки**
На основе ваших ответов я вижу такую структуру:

#### 1. **CLI-утилита**  
Компактная программа (на Python/Go/Rust), которая:  
- Принимает аргументы:  
  ```bash
  build-tracker \
    --token="SECRET_TOKEN" \
    --backend-url="https://your-backend/api/v1/events" \
    --project-id="web-app" \
    --commit-hash="a1b2c3d" \
    --status="started|in_progress|success|failed" \
    --log-file="/path/to/build.log"  # или --log-content="raw text"
  ```
- Отправляет JSON в бэкенд:  
  ```json
  {
    "project_id": "web-app",
    "commit_hash": "a1b2c3d",
    "status": "started",
    "timestamp": "2025-11-21T12:00:00Z",
    "logs": "base64-encoded-log-content"
  }
  ```
- Поддерживает retry-логику и аутентификацию.

#### 2. **Бэкенд**  
Сервис (на Golang), который:  
- Принимает события от CLI через REST.  
- Хранит данные в БД (PostgreSQL для метаданных, S3 для артефактов)
- Для хеша коммита запрашивает детали через Git API (Gitea).  
- Запускает асинхронные задачи через очередь (Celery/RQ/Temporal):  
  - Анализ изменений.  
  - Рассылка уведомлений (Slack/Email/Telegram).  
  - Отправка вебхуков в другие системы.  
  - Обновление контейнеров (интеграция с Portainer Webhok).  

#### 3. **Интеграция с CI/CD**  
Пример для GitLab CI (`.gitlab-ci.yml`):  
```yaml
stages:
  - build
  - notify
  - deploy

build_web_app:
  stage: build
  script:
    - build-tracker --token=$TRACKER_TOKEN --backend-url=$TRACKER_URL --project-id=web-app --commit-hash=$CI_COMMIT_SHA --status=started
    - npm run build
    - build-tracker --token=$TRACKER_TOKEN --backend-url=$TRACKER_URL --project-id=web-app --commit-hash=$CI_COMMIT_SHA --status=in_progress

deploy_to_prod:
  stage: deploy
  script:
    - build-tracker --token=$TRACKER_TOKEN --backend-url=$TRACKER_URL --project-id=web-app --commit-hash=$CI_COMMIT_SHA --status=deploy
```

---

### 🚀 **Шаг 3: План реализации**
Разобьем на этапы для AI-агента:

#### **Этап 1: CLI-утилита**  
- [ ] Выбрать язык (рекомендую **Go** для бинарника без зависимостей).  
- [ ] Реализовать парсер аргументов (библиотека `cobra` для Go).  
- [ ] Добавить HTTP-клиент с retry-логикой (библиотека `retryablehttp`).  
- [ ] Поддержать отправку логов как файла, так и текста.  
- [ ] Собрать статический бинарник для Linux/ARM64.  

#### **Этап 2: Бэкенд**  
- [ ] Настроить REST API на FastAPI (Python) с аутентификацией по токену.  
- [ ] Подключить PostgreSQL + SQLAlchemy ORM.  
- [ ] Интегрировать Celery + Redis для асинхронных задач.  
- [ ] Написать worker для:  
  - Запроса данных о коммите из Git.  
  - Формирования уведомлений.  
  - Отправки вебхуков.  
- [ ] Реализовать эндпоинт `/webhooks/update-containers` для ручного триггера.  

#### **Этап 3: Интеграция с CI/CD**  
- [ ] Добавить секреты (`TRACKER_TOKEN`, `TRACKER_URL`) в CI.  
- [ ] Вставить вызов `build-tracker` в ключевые стадии пайплайна.  
- [ ] Настроить артефакты для сохранения логов сборки.  

#### **Этап 4: Инфраструктура**  
- [ ] Задеплоить бэкенд в Docker-контейнере (docker-compose для старта).  
- [ ] Настроить reverse-proxy (Nginx).  
- [ ] Запустить Redis для очередей.  

---

### 💡 **Шаг 4: Пример кода для старта**
**CLI-утилита (Go):**  
```go
package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	token := flag.String("token", "", "Auth token")
	backendURL := flag.String("backend-url", "", "Backend API URL")
	projectID := flag.String("project-id", "", "Project ID")
	commitHash := flag.String("commit-hash", "", "Git commit hash")
	status := flag.String("status", "", "Build status")
	flag.Parse()

	payload := map[string]interface{}{
		"project_id":  *projectID,
		"commit_hash": *commitHash,
		"status":      *status,
	}

	client := http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", *backendURL, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+*token)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		log.Fatal("Failed to send event")
	}
}
```

**Бэкенд (FastAPI):**  
```python
from fastapi import FastAPI, Header, HTTPException
from pydantic import BaseModel

app = FastAPI()

class BuildEvent(BaseModel):
    project_id: str
    commit_hash: str
    status: str
    logs: str  # base64-encoded

@app.post("/api/v1/events")
async def handle_event(event: BuildEvent, authorization: str = Header(None)):
    if not authorization or "Bearer " not in authorization:
        raise HTTPException(status_code=401)
    
    token = authorization.split("Bearer ")[1]
    if token != "valid_token":  # В продакшене — проверка в БД/Redis
        raise HTTPException(status_code=403)
    
    # Сохранить в БД, запустить Celery task
    return {"status": "accepted"}
```