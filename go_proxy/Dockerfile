# Используем образ Golang для сборки приложения
FROM golang:1.17 AS builder

# Установка переменной окружения GO111MODULE в значение on, чтобы включить модули Go
ENV GO111MODULE=on

# Копируем все файлы в рабочую директорию
WORKDIR /app
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Отдельный этап контейнеризации
FROM alpine:latest

# Устанавливаем пакеты ca-certificates, чтобы обеспечить HTTPS соединение
RUN apk --no-cache add ca-certificates

# Копируем бинарный файл из предыдущего этапа
COPY --from=builder /app/main /main

# Команда, которая будет запускаться при старте контейнера
CMD ["/main"]
