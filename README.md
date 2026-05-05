# Кейс: сервис приема и обработки заявок

Проект демонстрирует полный цикл работы с пользовательскими заявками: публичная форма принимает обращение, backend сохраняет его в PostgreSQL, а администратор управляет заявками через защищенную админскую часть.

Основной фокус кейса - backend: API, хранение данных, авторизация администратора, сессии, фильтрация заявок и контейнеризация сервиса.

## Задача

Нужно было собрать небольшой, но законченный сервис для обработки лидов:

- пользователь оставляет заявку с именем, контактом, темой и описанием;
- заявка сохраняется в базе данных;
- администратор входит в систему;
- администратор видит список заявок, ищет и фильтрует их;
- администратор меняет статус обработки заявки;
- проект запускается одной командой через Docker Compose.

## Результат

Реализован сервис из трех контейнеров:

| Контейнер | Роль |
| --- | --- |
| `server` | Go backend API |
| `postgres` | PostgreSQL 17 |
| `frontend` | UI и nginx-прокси до API |

После запуска доступны:

- публичная форма: `http://localhost:5173`
- админка: `http://localhost:5173/admin`
- API: `http://localhost/api`
- healthcheck: `http://localhost/health`

## Пользовательские сценарии

### Создание заявки

Пользователь открывает публичную форму, заполняет данные и отправляет заявку. Backend валидирует обязательные поля, создает запись в таблице `leads` и возвращает ID новой заявки.

### Работа администратора

Администратор входит по логину и паролю. После успешной авторизации backend выдает пару токенов:

- `access` JWT для защищенных запросов;
- `refresh` token для продления сессии.

В админской части можно:

- просматривать заявки;
- искать по имени, теме, описанию и контакту;
- фильтровать по статусу и дате создания;
- сортировать список;
- менять статус заявки.

## Backend-решение

Backend написан на Go и разделен на слои:

```text
cmd/server.go         точка входа
internal/httpserver   router, handlers, middleware
internal/app          бизнес-логика
internal/storage      работа с PostgreSQL
internal/libs         конфиг и логирование
```

При старте приложение:

1. Читает переменные окружения из `.env`.
2. Создает JSON-логгеры для общего лога, HTTP и PostgreSQL.
3. Подключается к PostgreSQL с несколькими попытками.
4. Выполняет миграцию таблиц.
5. Создает дефолтного администратора, если его еще нет.
6. Запускает HTTP-сервер.

## Хранение данных

PostgreSQL используется как основное хранилище. Миграция выполняется автоматически при старте backend.

Создаются таблицы:

- `leads` - заявки пользователей;
- `admins` - учетные записи администраторов;
- `admin_sessions` - refresh-сессии.

Для ускорения выборок добавлены индексы по статусу и дате создания заявок, а также по полям сессий администратора.

## Авторизация

Админские маршруты защищены middleware, который проверяет заголовок:

```http
Authorization: Bearer <access>
```

Access token живет 30 минут. Refresh token живет 7 дней и хранится в базе. При обновлении токенов старая refresh-сессия удаляется, а новая создается заново.

Пароль администратора хранится в виде bcrypt-хеша.

## API

Публичные маршруты:

| Method | Path | Назначение |
| --- | --- | --- |
| `GET` | `/health` | Проверка доступности backend |
| `POST` | `/api/lead` | Создание заявки |

Админские маршруты:

| Method | Path | Назначение |
| --- | --- | --- |
| `POST` | `/api/admin/login` | Вход администратора |
| `POST` | `/api/admin/refresh` | Обновление токенов |
| `GET` | `/api/admin/leads` | Список заявок |
| `GET` | `/api/admin/leads/{id}` | Заявка по ID |
| `PATCH` | `/api/admin/leads/{id}/status` | Смена статуса |
| `POST` | `/api/admin/logout` | Выход и удаление refresh-сессии |

Список заявок поддерживает параметры:

- `status`
- `q`, `query`, `search`
- `date_from`
- `date_to`
- `limit`
- `offset`
- `sort`
- `order`

Подробная OpenAPI-схема находится в `backend-docs/swagger.yaml`.

## Инфраструктура

Проект контейнеризован через Docker Compose.

Backend собирается из `backend-api/Dockerfile`, frontend собирается как Vite SPA и отдается через nginx. Nginx также проксирует `/api` и `/health` в backend-контейнер `server:80`, поэтому frontend использует относительные API-пути.

Логи backend проброшены на хост:

```text
backend-logs/
```

## Конфигурация

Создайте .env конфиг и заполните своими данными
```env
# Настройка сервера
SERVER_ADDR="0.0.0.0:80"
SERVER_READ_TIMEOUT_SECOND=15
SERVER_WRITE_TIMEOUT_SECOND=15

# Настройка логов
SERVER_LOG_PATH="./logs/server.log"
SERVER_HTTP_LOG_PATH="./logs/server_http.log"
SERVER_POSTGRES_LOG_PATH="./logs/server_postgres.log"

# Настройка базы данных
POSTGRES_ADDR="postgres:5432"
POSTGRES_USER="admin"
POSTGRES_PASSWORD="admin"
POSTGRES_DB="demo"


# Настройка данных админисратора
SERVER_ADMIN_LOGIN="admin"
SERVER_ADMIN_PASSWORD="admin"
JWT_SECRET="change-me-demo-secret"


```

## Запуск

```powershell
docker compose up --build
```

Остановка:

```powershell
docker compose down
```

Остановка с удалением данных PostgreSQL:

```powershell
docker compose down -v
```

## Стек

- Go 1.26
- chi
- PostgreSQL 17
- JWT
- bcrypt
- Docker Compose
- React + Vite