FROM golang:1.21-alpine AS builder

# Настройка зеркал
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache git

# Настройка GOPROXY
ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go fmt ./... && \
    go vet ./... && \
    # Сборка с оптимизациями
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o stats-service ./cmd/main.go

# Финальный образ
FROM alpine:latest

# Только необходимые пакеты
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk --no-cache add ca-certificates && \
    rm -rf /var/cache/apk/*

WORKDIR /root/
COPY --from=builder /app/stats-service .

EXPOSE 8081
CMD ["./stats-service"]