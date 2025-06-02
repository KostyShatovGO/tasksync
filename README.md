Название проекта ("TaskSync").
Краткий обзор (например, "Сервис для синхронизации задач").
Структуру папок (cmd/, internal/, pkg/).
Инструкции: установка Go 1.24.0, Docker, запуск (go mod tidy, docker compose up -d).

## API Endpoints

- **POST /register**
  - Описание: Регистрация нового пользователя.
  - Тело запроса: `{"username": "string", "password": "string"}`.
  - Успешный ответ: `{"id": int, "username": "string"}` (статус 200).
  - Ошибки:
    - 400: Неверный запрос (например, пустые поля).
    - 409: Пользователь с таким именем уже существует.
    - 500: Внутренняя ошибка сервера.

- **POST /login**
  - Описание: Авторизация пользователя.
  - Тело запроса: `{"username": "string", "password": "string"}`.
  - Успешный ответ: `{"token": "string"}` (статус 200).
  - Ошибки:
    - 400: Неверный запрос.
    - 401: Неверный пароль.
    - 500: Внутренняя ошибка сервера.

    ## Запуск проекта

1. Установи Docker и Docker Compose.
2. Создай файл `.env` с переменной `DATABASE_URL` (например, `DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable`).
3. Выполни:
   ```bash
   docker compose up --build