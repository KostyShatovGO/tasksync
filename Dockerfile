# Базовый образ с Go (Debian-based)
FROM golang:1.24.0 AS builder

# Установка зависимостей
RUN apt-get update && apt-get install -y git

# Рабочая директория
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем остальные файлы
COPY . .

# Проверяем наличие файла main.go перед сборкой
RUN ls -la /app/cmd/api && test -f /app/cmd/api/main.go || (echo "Error: /app/cmd/api/main.go not found" && exit 1)

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /tasksync ./cmd/api/main.go

# Финальный образ (Alpine для минимального размера)
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /tasksync /app/

# Установка зависимостей в финальном образе
RUN apk add --no-cache ca-certificates

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./tasksync"]