FROM golang:1.21-alpine AS builder

# Настройка зеркал для Alpine в builder
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o stats-service ./cmd/main.go

FROM alpine:latest

# Используем зеркало Яндекса для установки ca-certificates
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/stats-service .
EXPOSE 8081
CMD ["./stats-service"]