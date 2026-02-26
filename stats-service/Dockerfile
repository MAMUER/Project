# Build stage
FROM golang:1.21-alpine AS builder

<<<<<<< HEAD
# Настройка зеркал для Alpine в builder
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache git
=======
# Добавляем DNS прямо в Dockerfile
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf
>>>>>>> 1e7b04020027f80335ecde6292f4c1be8a895687

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o stats-service ./cmd

# Final stage
FROM alpine:latest

<<<<<<< HEAD
# Используем зеркало Яндекса для установки ca-certificates
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
=======
# Используем российское зеркало (не требует DNS)
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
>>>>>>> 1e7b04020027f80335ecde6292f4c1be8a895687
    apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/stats-service .
EXPOSE 8082
CMD ["./stats-service"]
