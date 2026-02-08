.PHONY: build run test test-integration clean install lint help

# Build the CLI binary (produces bin/sl)
build:
	go build -o bin/sl cmd/main.go

# Run the CLI
run: build
	./bin/sl

# Run unit tests (excludes integration tests that require external tools)
test:
	go test -v ./pkg/... ./internal/... ./cmd/...

# Run integration tests (requires mise, beads, perles installed)
test-integration:
	go test -v ./tests/integration/...

# Run tests with coverage (excludes integration tests)
test-coverage:
	go test -v -coverprofile=coverage.out ./pkg/... ./internal/... ./cmd/...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run golangci-lint
lint:
	golangci-lint run ./...

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/sl-linux cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/sl-darwin cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/sl-windows.exe cmd/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/sl-linux-arm64 cmd/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/sl-darwin-arm64 cmd/main.go

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Install the CLI to $GOBIN or $GOPATH/bin
install: build
	@if [ -n "$(GOBIN)" ]; then \
		echo "Installing sl to $(GOBIN)/sl"; \
		mkdir -p $(GOBIN); \
		cp bin/sl $(GOBIN)/sl; \
		echo "✓ Installed successfully to $(GOBIN)/sl"; \
	else \
		INSTALL_DIR=$$(go env GOPATH)/bin; \
		echo "Installing sl to $$INSTALL_DIR/sl"; \
		mkdir -p $$INSTALL_DIR; \
		cp bin/sl $$INSTALL_DIR/sl; \
		echo "✓ Installed successfully to $$INSTALL_DIR/sl"; \
	fi

# Help
help:
	@echo "Available targets:"
	@echo "  make build            - Build the CLI binary (produces bin/sl)"
	@echo "  make install          - Build and install sl to $$GOBIN or $$GOPATH/bin"
	@echo "  make run              - Build and run the CLI"
	@echo "  make test             - Run unit tests (excludes integration tests)"
	@echo "  make test-integration - Run integration tests (requires external tools)"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make fmt              - Format Go code"
	@echo "  make vet              - Run go vet"
	@echo "  make lint             - Run golangci-lint"
	@echo "  make build-all        - Build for all platforms"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make help             - Show this help"
	@echo ""
	@echo "Installation:"
	@echo "  make install        - Install from source"
	@echo "  go install ./cmd/main.go@latest - Install with go toolchain"
