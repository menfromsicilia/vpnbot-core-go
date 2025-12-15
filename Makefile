.PHONY: build run test clean docker docker-build docker-run deps

# Build binary
build:
	go build -o vpnbot-core cmd/server/main.go

# Run locally
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Test with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f vpnbot-core
	rm -f *.db *.db-shm *.db-wal

# Install dependencies
deps:
	go mod download
	go mod tidy

# Docker build
docker-build:
	docker build -t vpnbot-core-go:latest .

# Docker run
docker-run:
	docker-compose up -d

# Docker stop
docker-stop:
	docker-compose down

# Docker logs
docker-logs:
	docker-compose logs -f

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Development with hot reload (requires air)
dev:
	air

# Build for production
build-prod:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -o vpnbot-core cmd/server/main.go

