# Pismo Transactions

REST service that manages cardholder accounts and their transactions. Built as a
technical case for a backend position at Pismo (Visa).

Written in Go 1.25 with MySQL 8.4. No HTTP framework, no ORM, no mock library —
standard library plus a MySQL driver, a decimal package, and `golang-migrate`.

---

## Quick start

You need Docker Desktop (with WSL2 backend on Windows) and `make`.

```bash
./run
```

That builds the API image, starts MySQL, waits for `/health` to respond, and
prints the endpoints. First run takes ~30s (image pull + build).

Manual alternative if you prefer not to use the script:

```bash
make up      # start
make down    # stop, keep DB data
make clean   # stop, wipe DB volume
make logs    # follow API logs
```

## API

| Method | Path                | Success | Errors     |
|--------|---------------------|---------|------------|
| POST   | `/accounts`         | 201     | 400, 409   |
| GET    | `/accounts/{id}`    | 200     | 400, 404   |
| POST   | `/transactions`     | 201     | 400, 422   |
| GET    | `/health`           | 200     | 503        |

Full contract: [`docs/openapi.yaml`](docs/openapi.yaml). Paste into
<https://editor.swagger.io> for a rendered view, or import it into Postman
(*New → Import → File*) for a ready-to-use collection.

### Examples

Create an account:

```bash
curl -s -X POST http://localhost:8080/accounts \
     -H 'Content-Type: application/json' \
     -d '{"document_number":"12345678900"}'
# 201 {"account_id":1,"document_number":"12345678900"}
```

Duplicate `document_number`:

```bash
curl -s -X POST http://localhost:8080/accounts \
     -H 'Content-Type: application/json' \
     -d '{"document_number":"12345678900"}'
# 409 {"error":{"code":"duplicate_account","message":"account already exists"}}
```

Get an account:

```bash
curl -s http://localhost:8080/accounts/1
# 200 {"account_id":1,"document_number":"12345678900"}
```

Create a purchase (server flips the sign to negative):

```bash
curl -s -X POST http://localhost:8080/transactions \
     -H 'Content-Type: application/json' \
     -d '{"account_id":1,"operation_type_id":1,"amount":"50.00"}'
# 201 {"transaction_id":1,"account_id":1,"operation_type_id":1,"amount":"-50","event_date":"..."}
```

Create a credit voucher (server keeps the sign positive):

```bash
curl -s -X POST http://localhost:8080/transactions \
     -H 'Content-Type: application/json' \
     -d '{"account_id":1,"operation_type_id":4,"amount":"60.00"}'
# 201 {"transaction_id":2,"account_id":1,"operation_type_id":4,"amount":"60","event_date":"..."}
```

Transaction referencing a missing account (returns 422, not 404 — the
`account_id` is a client input, not the primary resource being fetched):

```bash
curl -s -X POST http://localhost:8080/transactions \
     -H 'Content-Type: application/json' \
     -d '{"account_id":999,"operation_type_id":1,"amount":"10.00"}'
# 422 {"error":{"code":"account_not_found","message":"account does not exist"}}
```

Every error response uses the same envelope:

```json
{ "error": { "code": "snake_case_code", "message": "human readable" } }
```

## Architecture

```
    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌───────┐
    │ httpapi  │───▶│ service  │───▶│repository│───▶│ MySQL │
    └──────────┘    └──────────┘    └──────────┘    └───────┘
                          │                │
                          └────────┬───────┘
                                   ▼
                            ┌──────────┐
                            │  domain  │  (imports nothing internal)
                            └──────────┘
```

Rules the codebase follows:

- **`domain` imports nothing from the other layers.** It owns the business
  rules — most importantly, the transaction sign rule
  (`OperationType.Sign()`), which is the single source of truth.
- **Interfaces are declared where they are consumed.** `AccountRepository`
  lives in `service/`, not `repository/mysql/`. Dependencies point inward,
  and swapping the MySQL implementation for another backend requires zero
  changes to `service/`.
- **Errors are values.** Domain sentinels (`ErrAccountNotFound`,
  `ErrDuplicateAccount`, ...) propagate untouched to `httpapi`, which maps
  them to HTTP status codes in one place (`internal/httpapi/errors.go`).
- **Configuration is env-only** (`os.Getenv` with sane defaults). Compose
  injects overrides for the containerized run.

## Configuration

| Variable          | Default     | Notes                                       |
|-------------------|-------------|---------------------------------------------|
| `HTTP_ADDR`       | `:8080`     | address to bind                             |
| `DB_HOST`         | `localhost` | overridden to `db` inside the compose network |
| `DB_PORT`         | `3306`      |                                             |
| `DB_USER`         | `pismo`     |                                             |
| `DB_PASSWORD`     | `pismopw`   |                                             |
| `DB_NAME`         | `pismo`     |                                             |

MySQL container-side variables (`MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`,
`MYSQL_USER`, `MYSQL_PASSWORD`) mirror the same values. Copy `.env.example`
to `.env` for local overrides — `.env` is gitignored.

## Testing

Two suites, split intentionally:

```bash
make test              # unit + handler tests, no DB, <1s
make test-integration  # repository tests against real MySQL (needs `make up` first)
```

**Heads up:** `make test-integration` `DELETE`s the `accounts` and
`transactions` tables between tests. It reuses the dev database because the
case does not warrant a dedicated one. If you were poking around in Postman,
run `make clean && make up` afterwards to start fresh.

Coverage per package:

| Package                        | Coverage | Kind                         |
|--------------------------------|----------|------------------------------|
| `internal/domain`              | 100%     | unit                         |
| `internal/service`             | 100%     | unit + hand-written fakes    |
| `internal/httpapi`             | 98.7%    | `httptest` + fakes           |
| `internal/repository/mysql`    | 80.8%    | integration (build tag)      |

## Project structure

```
.
├── cmd/api/main.go             # wiring only: config → db → migrations → handlers → server
├── embed.go                    # //go:embed migrations/*.sql (root package required)
├── migrations/                 # SQL up/down files, applied at startup
├── docs/
│   ├── case.md                 # original tech case
│   └── openapi.yaml            # HTTP contract
├── internal/
│   ├── config/                 # env-based config
│   ├── db/                     # sql.DB open + connection retry
│   ├── domain/                 # entities, sign rule, sentinel errors — no external deps
│   ├── httpapi/                # handlers, DTOs, error mapping, access-log middleware
│   ├── migrate/                # golang-migrate runner over embed.FS
│   ├── repository/mysql/       # SQL implementation + integration tests
│   └── service/                # use cases; defines repository interfaces
├── Dockerfile                  # multi-stage, distroless, non-root
├── docker-compose.yml          # api + mysql with healthcheck
├── Makefile                    # up/down/clean/run/test/logs/build
└── run                         # ./run — one-command boot
```

## Design decisions worth calling out

- **`net/http` standard library, no framework.** Go 1.22+ ServeMux does
  method-aware routing (`"POST /accounts"`) and path parameters (`{id}`).
  Chi/Gin buy nothing here for four endpoints.
- **`shopspring/decimal`, never `float64`, for money.** `DECIMAL(15,2)` in
  MySQL, string-encoded in JSON. Zero float precision surprises.
- **Sign rule lives in `OperationType.Sign()` — one function, one place.**
  Adding a fifth operation type is one `case` in that method plus one seed
  row. Neither the service nor the handler needs to change.
- **Interfaces declared where they are consumed.** `AccountRepository` in
  `service/`, `Pinger` in `httpapi/`. Concrete implementations satisfy them
  implicitly — no `implements` keyword, no boilerplate.
- **Distroless non-root runtime.** No shell, no package manager, no `apt`.
  Attack surface reduced to the compiled binary and libc.
- **`DisallowUnknownFields` + `MaxBytesReader(1 MiB)` on every JSON
  decoder.** Silent field drops are a classic contract-drift bug; unbounded
  request bodies are trivial DoS.
- **Fakes hand-written in `_test.go`, not generated.** Interfaces here have
  1–3 methods; a fake fits in ~15 lines. Skipping `gomock`/`mockery` keeps
  tests readable at a glance.
- **Integration tests behind a build tag.** `go test ./...` stays fast and
  DB-free; `go test -tags=integration ./...` exercises the real MySQL.

## Next steps (production hardening beyond the case scope)

- **AWS deployment:** ECS Fargate + RDS MySQL + Secrets Manager for DB
  credentials, ALB in front.
- **CI/CD:** GitHub Actions running `make lint` + `make test` +
  `make test-integration` (against a MySQL service container) on every PR.
- **HTTP hardening:** `ReadHeaderTimeout`, `WriteTimeout`, `IdleTimeout`,
  and graceful shutdown on SIGTERM. Planned in the next commits.
- **Security lint pipeline:** `gosec` + `govulncheck` gating merges. Planned
  in the next commits.
- **Observability:** `/metrics` endpoint scraped by Prometheus with a Grafana
  dashboard. Planned as a stretch goal in this repo.
- **Rate limiting** at the ALB, or a Redis-backed leaky-bucket in front of
  `POST /transactions`.

## Requirements

- Docker Desktop with WSL2 (Windows) or native Docker (macOS/Linux)
- `make`
- (Optional, for running Go tooling on the host) Go 1.25+

## License

MIT
