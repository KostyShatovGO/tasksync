# Базовый образ с Go
FROM golang:1.22-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /tasksync

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /tasksync /app/tasksync
COPY --from=builder /app/configs /app/configs

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./tasksync"]