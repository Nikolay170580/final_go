# Используем многостадийную сборку
FROM golang:1.25.0-alpine AS builder

# Устанавливаем зависимости для сборки (CGO требует gcc/musl-dev)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости (кэшируется отдельно)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и компилируем
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM alpine:3.20

# Базовые зависимости для работы
RUN apk --no-cache add ca-certificates

# Создаем непривилегированного пользователя
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Копируем бинарник и статику из builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web

# Создаем директорию для БД и назначаем права
RUN mkdir -p /data && chown -R app:app /data

USER app

# Переменные окружения по умолчанию
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db

# Запускаем приложение
CMD ["./main"]