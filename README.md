# Automatic Message Sending System

An **Automatic Message Sending System** that periodically sends SMS messages through a webhook-based gateway, stores message statuses in PostgreSQL, and exposes a REST API for monitoring and querying sent messages.

---

## Supported Environment

- Go 1.25+
- PostgreSQL 15+
- Docker & Docker Compose
- Linux, macOS, or Windows (with Docker)

---

## Building the Project

Clone the repository and build the binary using Go:

```bash
# using SSH
git clone git@github.com:lazerion/outbox-relayer.git
cd outbox-relayer
go mod download
go build -o outbox-relayer ./cmd/main.go
```

Alternatively, build a Docker image:

```bash 
docker build -t outbox-relayer:local .
```

## Running Tests
Unit and integration tests can be executed using:

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...
```

Integration tests use Testcontainers for PostgreSQL, so no manual database setup is required.

## Docker Usage
Start the application and PostgreSQL using Docker Compose:

```bash
docker-compose up --build
```

This will spin up:
- postgres container with a devdb database
- outbox-relayer container exposing port 8080

Alternatively, run manually with a local PostgreSQL instance:
```bash
docker run --rm -it -p 8080:8080 \
  -e POSTGRES_HOST=host.docker.internal \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=dev \
  -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=devdb \
  outbox-relayer:local
```

Configuration is loaded from internal/config/config.yaml and can be overridden using environment variables.

## Swagger Endpoint

The API is documented with Swagger:

Endpoints include:

- GET /messages/sent – Query sent messages with cursor-based pagination
- POST /scheduler/toggle – Start/stop message sending scheduler

## Sender Response Handling

The sender accepts HTTP 202 responses from the gateway. A typical accepted message response looks like:

```json
{
  "message": "Accepted",
  "messageId": "67f2f8a8-ea58-4ed0-a6f9-ff217df4d849"
}
```

## Error Handling

The `RelayerService` implements robust error handling with transactional safety:

- **Transactional Safety:** All pending message updates occur within a single database transaction. Messages are marked `sent`, `failed`, or have their attempt count incremented atomically. If any error occurs before committing, the transaction is rolled back.

- **Recoverable vs Unrecoverable Errors:**
    - Recoverable errors (e.g., temporary network issues) increment the message attempt count to retry later.
    - Unrecoverable errors (e.g., invalid payload) mark the message as `failed` immediately.

- **Upstream Response Handling:**
    - Messages accepted by the upstream gateway (`"accepted"`) are marked as `sent`.
    - Any other gateway response marks the message as `failed` with logged details.

This ensures that each message is processed safely, and failures do not leave the system in an inconsistent state.

### Database Migration Support

The system uses [golang-migrate](https://github.com/golang-migrate/migrate) to handle database schema migrations.

- Migration files are located in the `internal/migrations` directory.
- Migrations are applied automatically on application startup via the `RunMigrations` function.
- The migration module is integrated with the lifecycle of the application using `fx.Lifecycle`.

This ensures the database schema is always up-to-date before the application starts processing messages.

## Improvements / Future Work
- Metrics Collection & Dashboard – Expose Prometheus metrics for message throughput, failures, and scheduler status.
- Alerting – Integrate with alerting systems (e.g., Slack, email) for failed message delivery.
- Retry Strategy Enhancements – Implement exponential backoff or dynamic scheduling.
- Implement Kubernetes-native Liveness and Readiness probes to manage the scheduler's lifecycle and traffic routing effectively.
- Implement a dedicated, read-only database (the "Query Store") separate from the transactional Write database

### Failed Messages Recovery / Replay
1. **Retry Queue with Backoff:**
    - Instead of max. attempt based failure, push messages with recoverable errors into a retry queue.
    - Use exponential backoff or fixed delay between retries to avoid spamming the gateway.

2. **Dead Letter / Replay Mechanism:**
    - Maintain a `failed_messages` table or status flag to track messages that exceeded retry attempts.
    - Provide an admin or automated process to manually or programmatically replay these messages after fixing issues (e.g., correcting invalid phone numbers or gateway downtime).