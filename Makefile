.PHONY: build run test clean help

# Build the CLI binary
build:
	go build -o bin/specledger cmd/main.go

# Run the CLI
run: build
	./bin/specledger

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/specledger-linux cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/specledger-darwin cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/specledger-windows.exe cmd/main.go

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the CLI binary"
	@echo "  make run            - Build and run the CLI"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make fmt            - Format Go code"
	@echo "  make vet            - Run go vet"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make help           - Show this help"
