FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем go.mod и go.sum из корня проекта (контекст сборки = ../..)
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник (путь относительно корня проекта)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/auth-service ./cmd/auth-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/auth-service .

USER appuser

EXPOSE 50051

CMD ["./auth-service"]