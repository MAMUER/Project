FROM golang:1.21-alpine AS builder

# Настройка зеркал для Alpine и Go
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache git

# Настройка GOPROXY для Go modules
ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on
ENV GOPRIVATE=*.company.com

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Проверка форматирования (fmt)
RUN test -z "$(go fmt ./...)" || (echo "Files are not properly formatted. Run 'go fmt' and commit changes." && exit 1)

# Проверка линтера (vet)
RUN go vet ./...

# Сборка приложения
RUN go build -o stats-service ./cmd/main.go

FROM alpine:latest

# Используем зеркало Яндекса
RUN echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirror.yandex.ru/mirrors/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk update && \
    apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/stats-service .
EXPOSE 8081
CMD ["./stats-service"]