# syntax=docker/dockerfile:1

# ========== Builder stage ==========
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/doctor-service/ ./cmd/doctor-service/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY api/ ./api/

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/doctor-service \
    ./cmd/doctor-service

# ========== Runtime stage ==========
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/doctor-service .

USER appuser

EXPOSE 50054

ENTRYPOINT ["./doctor-service"]
