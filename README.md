# GoQuizVibe
Платформа для обучения и проведения викторин

## Stack
**Backend**
- Go, `net/http` + `http.ServeMux` (method-prefixed routing)
- [Templ](https://templ.guide/) — type-safe HTML-шаблоны
- [HTMX](https://htmx.org/) — интерактивность без SPA
- Tailwind CSS + [rustywind](https://github.com/avencera/rustywind) (сортировка классов)

**Хранение и инфраструктура**
- PostgreSQL 18 + [sqlc](https://sqlc.dev/) (type-safe SQL → Go), миграции — `golang-migrate`
- MinIO (S3-compatible) — хранение изображений к вопросам
- Redis 8 — кэш, сессии викторин, таймеры
- Nginx — reverse proxy

**Микросервисы**
- `typst` — gRPC-сервис компиляции Typst-разметки в изображения (с интеграцией с MinIO)

**Auth & конфигурация**
- JWT в HttpOnly cookie (stateless auth)
- `.env` → `config/config.go`
- [Wire](https://github.com/google/wire) — внедрение зависимостей

**Локализация**
- `gettextgocodegen` — генерация type-safe переводов из `.po` файлов (ru/en)

**Наблюдаемость**
- Prometheus + Grafana + node-exporter
- Adminer — веб-интерфейс для PostgreSQL

**Контейнеризация**
- podman-compose / docker-compose

## Features
- 👥 **Роли:** `teacher` и `student` с разделением доступа
- 📝 **Викторины** с типами вопросов: `choice` (одиночный/множественный выбор), `open` (свободный ответ), `fill` (заполнение пропусков)
- 🖼 **Изображения к вопросам** — загрузка в MinIO с предпросмотром в админке
- 🏆 **Геймификация** — прогресс, достижения, мотивационные механики (`feature/gamification`)
- 📚 **Учебные материалы** (`feature/learning_materials`) — привязка материалов к викторинам
- ✏️ **Typst-редактор** — компиляция Typst-разметки в изображения (gRPC-микросервис)
- 📊 **Дашборд** студента и учителя с историей попыток и статистикой
- 🌍 **Локализация** интерфейса (ru/en) с авто-определением по `Accept-Language`
- ⚡ **HTMX-обновления** — partial HTML без перезагрузки страницы
- 🛡 **Stateless JWT-auth** в HttpOnly cookie
- 🚦 **Мониторинг** — метрики Prometheus, дашборды Grafana

## Запуск через docker-compose

### 1. Подготовка окружения
```bash
# Клонирование
git clone https://github.com/Glootea/GoQuizVibe.git
cd GoQuizVibe

# Перенаправление трафика на MinIO (обязательно)
echo "127.0.0.1 minio" | sudo tee -a /etc/hosts
```

### 2. Конфигурация `.env`
Файл `deployment/.env` уже содержит рабочие моковые значения по умолчанию. При необходимости отредактируйте пароли и хосты.

### 3. Запуск инфраструктуры (БД, MinIO, Redis, Typst, мониторинг, Nginx)
```bash
cd deployment
podman compose up -d
# или: docker compose up -d
```

Поднимутся контейнеры:
| Сервис       | Порт  | Назначение                                  |
|--------------|-------|---------------------------------------------|
| `db`         | 5432  | PostgreSQL 18                               |
| `minio`      | 9000  | S3 API                                      |
| `minio`      | 9001  | MinIO Console (http://localhost:9001)       |
| `redis`      | 6379  | Кэш/сессии                                  |
| `typst`      | 9091  | gRPC-микросервис Typst (→ 9090 в контейнере)|
| `adminer`    | 8081  | Веб-интерфейс PostgreSQL                    |
| `prometheus` | 9090  | Сбор метрик                                 |
| `grafana`    | 3000  | Дашборды (admin / admin123)                 |
| `nginx`      | 80    | Reverse proxy                               |

### 4. Запуск backend (приложение)
Сервис `server` в compose помечен профилем `with-server` — по умолчанию не стартует. В dev-режиме backend запускается локально:
```bash
cd scripts
make dev        # dc-up + генерация шаблонов/locales + watch
# или
make generate   # одноразовая генерация всех ассетов
make run        # запуск без watcher'ов
```

Если хотите собрать backend в контейнере:
```bash
cd deployment
podman compose --profile with-server up -d --build server
# сервер будет доступен на http://localhost:7890
```

### 5. Полезные команды
```bash
# Остановить всё
make dc-down
# или
cd deployment && podman compose down

# Только мониторинг
make monitoring-up
make monitoring-down

# Полная очистка (с удалением томов)
cd deployment && podman compose down -v
```

Важно: для работы minio нужно перенаправить трафик (в /etc/hosts (macos/linux) вставить `127.0.0.1 minio`)

## Архитектурные решения
### HTML-first с HTMX
Фронтенд построен на **Templ** (Go-шаблонизатор с type-safety) в сочетании с **HTMX** для интерактивности без SPA-сложности.

### SQL-first база данных
**sqlc** генерирует type-safe Go-код из SQL-запросов. Запросы живут в `sql/queries/*.sql`, код генерируется в `db/*.sql.go`.

### Custom errors с HTTP status
Ошибки несут HTTP-статус через интерфейс `HTTPStatus() int`:
```go
// custom_errors/custom_errors.go:17
func WithHTTPStatus(err error, status int) error {
    return statusError{err: err, status: status}
}
```
Это позволяет единообразно обрабатывать ошибки:
```go
// handlers/quiz.go:36
return ce.WithHTTPStatus(errors.Join(ce.ErrNotFound, err), http.StatusNotFound)
```
### Stateless auth с JWT в cookie
JWT хранится в HttpOnly cookie (`token`). На каждом запросе middleware валидирует токен.

### S3-compatible storage для изображений
MinIO используется для хранения изображений вопросов. Это позволяет легко переключиться на S3 в продакшене.


## Локализация
Используется **gettextgocodegen** для генерации type-safe функций перевода из `.po` файлов.


