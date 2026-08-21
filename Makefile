include .env
.PHONY: build run test test-unit lint fmt vet swagger tidy clean coverage check migration migrate-up seed

# Build
build:
	go build -o bin/workout-tracker ./cmd/api/...

# Run
run:
	go run ./cmd/api/...

# Unit tests (all packages, race detector)
test-unit:
	go test -race -count=1 ./...

# Format
fmt:
	gofmt -s -w -s -w .

# Vet
vet:
	go vet ./...

# Lint
lint:
	go vet ./...

# Swagger
swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Dependency management
tidy:
	go mod tidy

# Coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html to view report"

# Generate swagger + run tests
check: swagger test-unit lint
	@echo "All checks passed"

# Create a new Goose migration
migration:
	goose -dir=migrations create $(filter-out $@,$(MAKECMDGOALS)) sql

# Database migrations
migrate-up:
	goose -dir=migrations postgres "$$DATABASE_URL" up

# Seed database
seed:
	go run ./cmd/seed/main.go

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html docs/