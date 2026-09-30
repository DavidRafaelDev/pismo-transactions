.PHONY: up down clean run test test-integration build logs

up: ## start API + MySQL via docker compose
	docker compose up -d --build

down: ## stop and remove containers (keeps DB volume)
	docker compose down

clean: ## stop and remove containers AND DB volume (destroys data)
	docker compose down -v

run: ## run the API locally (requires MySQL from compose)
	go run ./cmd/api

test: ## run unit and handler tests (no DB required)
	go test ./...

test-integration: ## run repository tests against MySQL (requires `make up` first — WIPES accounts/transactions)
	go test -tags=integration -cover -v ./internal/repository/mysql/...

build: ## build the API binary to bin/api
	go build -o bin/api ./cmd/api

logs: ## follow API container logs
	docker compose logs -f api
