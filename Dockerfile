# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies for cgo if needed
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o outbox-relayer ./cmd/server/main.go

# Stage 2: Run
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/outbox-relayer .

COPY internal/infra/migrations ./internal/infra/migrations

# Copy config file
COPY internal/config/config.yaml ./internal/config/config.yaml

COPY internal/api/docs ./internal/api/docs

# Expose API port
EXPOSE 8080

CMD ["./outbox-relayer"]
