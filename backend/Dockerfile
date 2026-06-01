# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Скачиваем зависимости отдельным слоем для кеша
COPY go.mod go.sum* ./
RUN go mod download

# Копируем исходники и собираем
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /api /app/api
COPY --from=builder /build/migrations /app/migrations

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/api"]
