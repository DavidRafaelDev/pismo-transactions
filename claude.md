# CLAUDE.md

Technical context for this repository. Full problem statement: @docs/case.md

## Project

A small REST service that manages cardholder **accounts** and their **transactions**.
Evaluation criteria, in order: **maintainability, simplicity, testability, documentation**.
Bonus points: Docker, a one-command `./run`, good docs, good tests.

Simplicity wins ties. Do not add frameworks, layers or abstractions the problem does not need.

## Domain rules

- An account has an `account_id` and a unique `document_number`.
- Every transaction belongs to one account and has one operation type:

  | ID | Description                | Stored sign |
  |----|----------------------------|-------------|
  | 1  | Normal Purchase            | negative    |
  | 2  | Purchase with installments | negative    |
  | 3  | Withdrawal                 | negative    |
  | 4  | Credit Voucher             | positive    |

- Clients always send a **positive** `amount`. The service applies the sign based on the
  operation type. This rule lives in the domain layer, in one place only.
- `event_date` is set by the server, in UTC, at creation time.
- Money is never a float: `DECIMAL(15,2)` in MySQL and `shopspring/decimal` in Go.
  Amounts must be `> 0` and have at most 2 decimal places.

## API contract

| Method | Path                   | Success | Errors                                                       |
|--------|------------------------|---------|--------------------------------------------------------------|
| POST   | `/accounts`            | 201     | 400 invalid body, 409 duplicate `document_number`            |
| GET    | `/accounts/{id}`       | 200     | 400 invalid id, 404 not found                                |
| POST   | `/transactions`        | 201     | 400 invalid body or operation type, 422 account not found    |
| GET    | `/health`              | 200     | 503 when the database is unreachable                         |
| GET    | `/metrics`             | 200     | Prometheus format                                            |

- JSON fields use `snake_case`, matching the case document.
- `POST /transactions` returns the created transaction, including `transaction_id`,
  the signed `amount` and `event_date`.
- Error body, always: `{"error": {"code": "snake_case_code", "message": "human readable"}}`.
- Keep `docs/openapi.yaml` in sync with any contract change.

## Stack and decisions

| Topic         | Choice                                                   |
|---------------|----------------------------------------------------------|
| Language      | Go (latest stable), modules                              |
| HTTP          | Standard library `net/http` with method + path patterns  |
| Database      | MySQL 8, `database/sql` + `go-sql-driver/mysql`, no ORM  |
| Migrations    | `golang-migrate`, SQL files embedded, run on startup     |
| Money         | `shopspring/decimal`                                     |
| Logging       | `log/slog`, JSON output                                  |
| Config        | Environment variables only, with sane local defaults     |
| Metrics       | `prometheus/client_golang`                               |
| Container     | Multi-stage Dockerfile, distroless, non-root user        |

Ask before adding any dependency not listed here.

## Layout

```
cmd/api/main.go          # wiring only: config, db, migrations, router, server
internal/
  domain/                # entities, operation types, sign rule, domain errors
  service/               # use cases; depends on repository interfaces
  repository/mysql/      # SQL implementations of the repository interfaces
  httpapi/               # handlers, router, middleware, request/response DTOs
  config/                # env parsing
migrations/              # *.up.sql / *.down.sql, including operation type seed
docs/                    # case.md, openapi.yaml
deploy/observability/    # Prometheus + Grafana provisioning (optional profile)
```

Dependency direction: `httpapi → service → domain`, and `repository → domain`.
The domain package imports nothing from the other layers.

## Conventions

- `gofmt` and `go vet` clean. Idiomatic Go over patterns borrowed from other languages.
- Constructor functions for dependency injection; no global mutable state.
- Define interfaces where they are consumed (in `service`), keep them small.
- Wrap errors with `%w`; map domain errors to HTTP status in one place in `httpapi`.
- Pass `context.Context` through every layer down to the database calls.
- Handlers only decode, validate shape, call the service and encode. No business rules.
- Small, focused commits with clear messages.

## Security

- Decode JSON with `DisallowUnknownFields` and a 1 MB body limit.
- Validate every input: `document_number` digits only (11 or 14 chars), ids positive,
  amount rules above.
- Parameterized queries only.
- Unique index on `document_number`; translate the duplicate-key error to 409.
- HTTP server with read, write and idle timeouts; graceful shutdown on SIGTERM.
- No secrets in code or images; credentials come from the environment.
- Mask `document_number` in logs.
- `gosec` and `govulncheck` run in `make lint`.

## Testing

- Unit tests for `domain` and `service`, using hand-written fakes for repositories.
- Handler tests with `net/http/httptest`, covering success and every error status.
- Repository integration tests against a real MySQL, behind the `integration` build tag.
- Table-driven tests. Aim for high coverage of `domain`, `service` and `httpapi`.

## Commands

| Command                  | What it does                                       |
|--------------------------|----------------------------------------------------|
| `./run`                  | Builds and starts API + MySQL (wraps `make up`)    |
| `make up` / `make down`  | Start / stop the Docker Compose stack              |
| `make run`               | Run the API locally against the Compose MySQL      |
| `make test`              | Unit + handler tests                               |
| `make test-integration`  | Repository tests against MySQL                     |
| `make cover`             | Coverage report                                    |
| `make lint`              | `go vet`, `gosec`, `govulncheck`                   |
| `make observability`     | Start Prometheus + Grafana with a ready dashboard   |

After any code change: run `make test` and `make lint` before considering the task done.
Update the README whenever a command, env var or endpoint changes.