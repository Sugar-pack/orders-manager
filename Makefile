.PHONY: docker-run
docker-run: vet lint test
	@docker-compose up --build -d --remove-orphans

.PHONY: docker-up
docker-up:
	@docker-compose up -d

.PHONY: docker-build
docker-build:
	@docker-compose build

vet:  ## Run go vet
	go vet ./...

lint: ## Run go lint
	golangci-lint run

test: ## Run tests
	go test ./...

test-coverage: ## Run go test with coverage
	go test ./... -coverprofile=coverage.out `go list ./...`