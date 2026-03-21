FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Генерация protobuf (требует protoc)
RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN ./scripts/generate-proto.sh

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/auth-service ./cmd/auth-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/auth-service .
COPY --from=builder --chown=appuser:appgroup /app/proto/gen ./proto/gen

USER appuser

EXPOSE 50051

CMD ["./auth-service"]