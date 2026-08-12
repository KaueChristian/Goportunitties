.PHONY: help dev api worker ingest web build test test-go test-web lint fmt cover docker clean

VERSION ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
BINARY  ?= goportunitties

help: ## Show this help
	@grep -E '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

api: ## Run the API on :8080
	go run ./cmd/api

worker: ## Run the ingestion worker on its configured schedule
	go run ./cmd/worker

ingest: ## Run one ingestion pass and exit
	INGESTION_INTERVAL=0 go run ./cmd/worker

web: ## Run the Vite dev server on :5173
	cd frontend && npm run dev

build: ## Build the single binary with the frontend embedded
	cd frontend && npm ci && npm run build
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/api
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY)-worker ./cmd/worker

test: test-go test-web ## Run every test

test-go: ## Run the Go tests
	go test ./... -race -cover

test-web: ## Run the frontend tests
	cd frontend && npm test

lint: ## Vet the Go code and lint the frontend
	gofmt -l ./cmd ./internal
	go vet ./...
	cd frontend && npm run lint

fmt: ## Format the Go code
	gofmt -w ./cmd ./internal

cover: ## Write an HTML coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "report at coverage.html"

docker: ## Build and start the container
	docker compose up --build

clean: ## Remove build artefacts
	rm -f $(BINARY) $(BINARY)-worker coverage.out coverage.html
	rm -rf internal/web/dist/assets internal/web/dist/index.html
