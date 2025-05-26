FROM golang:1.23-alpine AS builder
WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o webscraper main.go

# Финальный минимальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/webscraper .

# Запускаем приложение
CMD ["./webscraper"]