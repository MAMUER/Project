FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates protobuf protobuf-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api-gateway ./cmd/api-gateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=appuser:appgroup /app/bin/api-gateway .

USER appuser

EXPOSE 8080

CMD ["./api-gateway"]