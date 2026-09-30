.PHONY: up down clean run test build logs

up: ## start API + MySQL via docker compose
	docker compose up -d --build

down: ## stop and remove containers (keeps DB volume)
	docker compose down

clean: ## stop and remove containers AND DB volume (destroys data)
	docker compose down -v

run: ## run the API locally (requires MySQL from compose)
	go run ./cmd/api

test: ## run unit and handler tests
	go test ./...

build: ## build the API binary to bin/api
	go build -o bin/api ./cmd/api

logs: ## follow API container logs
	docker compose logs -f api
