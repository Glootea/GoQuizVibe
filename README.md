# GoQuizVibe
Интерактивная платформа для проведения викторин с геймификацией.
## Архитектурные решения
### HTML-first с HTMX
Фронтенд построен на **Templ** (Go-шаблонизатор с type-safety) в сочетании с **HTMX** для интерактивности без SPA-сложности.
**Почему:** Простой проект — простой стек. Не нужен React/Vue для CRUD-приложения с парой интерактивных элементов. HTMX позволяет обновлять отдельные части страницы (`hx-target`, `hx-swap`) без перезагрузки.
**Как работает:**
- Первая загрузка страницы — полный HTML через Templ
- Последующие взаимодействия (ответ на вопрос, навигация) — HTMX-requests, возвращающие partial HTML
- Сервер определяет HTMX по заголовку `HX-Request` и решает, какой HTML отдать (полная страница или partial)
```go
// handlers/quiz.go:83
isHtmx := r.Header.Get("HX-Request") == "true"
if !isHtmx {
    return pages.QuizPage(data).Render(r.Context(), w)
}
return pages.QuestionCard(quiz, index, sessionIDStr).Render(r.Context(), w)
```
### SQL-first база данных
**sqlc** генерирует type-safe Go-код из SQL-запросов. Запросы живут в `sql/queries/*.sql`, код генерируется в `db/*.sql.go`.
**Почему:** 
- SQL проще читать и поддерживать, чем ORM-построение цепочек
- Генерённый код гарантирует соответствие запросов схеме
- Нет проблемы N+1 — запросы явные
**Структура:**
```
sql/
├── migrations/     # Миграции (golang-migrate)
└── queries/       # SQL для sqlc
```
### Разделение ответственности
```
handlers/    → HTTP-логика (парсинг, валидация, ответ)
services/    → Бизнес-логика (quin, session, auth)
db/          → Уровень данных (сгенерирован sqlc)
pages/       → Templ-компоненты (представление)
```
**Принцип:** Handlers не содержат логики. Они получают данные из сервисов и рендерят страницы. Сервисы не знают о HTTP.
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
### JSONB для гибких данных
Колонки `options` (вопросы), `answers` (сессии) хранят JSONB для гибкости без схемы:
```go
// services/quiz.go:99
if err := json.Unmarshal(session.Answers, &answers); err != nil { ... }
```
### Stateless auth с JWT в cookie
JWT хранится в HttpOnly cookie (`token`). На каждом запросе middleware валидирует токен.
**Почему не сессии:** Простота. Не нужен Redis для sessions. Минус — нельзя "выйти из всех устройств".
### S3-compatible storage для изображений
MinIO используется для хранения изображений вопросов. Это позволяет легко переключиться на S3 в продакшене.
```go
// services/admin.go:264
url, err := s.storageService.UploadImage(ctx, fileHeader)
```
## Структура проекта
```
├── main.go                    # Инициализация, роутинг
├── config/                    # Загрузка из .env
├── database/                  # Подключение, миграции, сиды
├── db/                        # sqlc-generated код
│   └── models.go              # Enum типы
├── models/                    # Доменные модели
├── handlers/                  # HTTP обработчики
├── services/                  # Бизнес-логика
├── pages/                     # Templ компоненты (.templ файлы)
├── middleware/                # Auth, Role, Compression
├── custom_errors/             # Ошибки с HTTP status
└── types/                     # Data transfer objects для страниц
```
## База данных
### Сущности
| Таблица | Назначение |
|---------|------------|
| `users` | Учителя и студенты (роль — enum) |
| `quizzes` | Викторины (статус — enum) |
| `questions` | Вопросы с типами choice/open/fill |
| `quiz_attempts` | Попытки прохождения |
| `user_answers` | Ответы пользователя |
| `quiz_sessions` | Активные сессии (current_index, answers JSONB) |
| `question_images` | Изображения вопросов |
### Enums
```sql
role = 'teacher' | 'student'
quiz_status = 'available' | 'assigned' | 'completed' | 'archived'
question_type = 'choice' | 'open' | 'fill'
```
## API дизайн
### REST-простой роутинг
Используется `http.ServeMux` с методом-префиксом (`GET /path`, `POST /path`).
**Простая админка:** Нет CRUD-паттернов — каждый ресурс имеет свои endpoints (`/admin/quizzes/new`, `/admin/quizzes/{id}/question`).
### Обработка ошибок
```go
// ErrorHandler统一处理所有 ошибки
wrapHandler := func(handler any) http.HandlerFunc {
    switch h := handler.(type) {
    case func(w http.ResponseWriter, r *http.Request) error:
        return handlers.ErrorHandler(h)
    }
}
```
### HTMX-ответы
Админка возвращает partial HTML для HTMX-targets:
```go
if r.Header.Get("hx-request") == "true" {
    return admin.QuestionsSection(quizWithQuestions, questions).Render(r.Context(), w)
}
```
## Запуск
```bash
# Генерация кода/html шаблонов
go tool templ generate
# Запуск
go run .
```

Также Makefile содержит полезные dev команды
