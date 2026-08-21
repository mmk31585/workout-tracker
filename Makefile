.PHONY: build run test test-unit lint fmt vet swagger tidy clean coverage check

# Build
build:
	go build -o bin/url-shortener ./cmd/api/...

# Run
run:
	go run ./cmd/api/...

# Unit tests (all packages, race detector)
test-unit:
	go test -race -count=1 ./...

# Format
fmt:
	gofmt -s -w .

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

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html docs/
