# Build stage
FROM golang:1.21-alpine AS builder

# Добавляем DNS прямо в Dockerfile
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o stats-service ./cmd

# Final stage
FROM alpine:latest

# Используем российское зеркало (не требует DNS)
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/stats-service .
EXPOSE 8082
CMD ["./stats-service"]
