.PHONY: build test clean install run fmt lint help

# Variables
BINARY_NAME=autobox
MAIN_PATH=main.go
BUILD_DIR=./bin
COVERAGE_FILE=coverage.out

# Build variables
VERSION?=0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X 'github.com/Autobox-AI/autobox-cli/cmd.Version=${VERSION}' \
	-X 'github.com/Autobox-AI/autobox-cli/cmd.BuildTime=${BUILD_TIME}' \
	-X 'github.com/Autobox-AI/autobox-cli/cmd.GitCommit=${GIT_COMMIT}'"

# Default target
all: build

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## //' | sort

## build: Build the binary
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ${MAIN_PATH}
	@echo "Binary built at ${BUILD_DIR}/${BINARY_NAME}"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=${COVERAGE_FILE} ./...
	@go tool cover -html=${COVERAGE_FILE} -o coverage.html
	@echo "Coverage report generated at coverage.html"

## install: Install the binary to GOPATH/bin
install: build
	@echo "Installing ${BINARY_NAME}..."
	@cp ${BUILD_DIR}/${BINARY_NAME} $$(go env GOPATH)/bin/${BINARY_NAME}
	@# Re-sign the binary to avoid macOS security issues
	@if [ "$$(uname)" = "Darwin" ]; then \
		echo "Re-signing binary for macOS..."; \
		codesign --force --deep --sign - $$(go env GOPATH)/bin/${BINARY_NAME}; \
	fi
	@echo "Installed to $$(go env GOPATH)/bin/${BINARY_NAME}"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf ${BUILD_DIR}
	@rm -f ${COVERAGE_FILE} coverage.html
	@echo "Clean complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

## run: Run the CLI
run: build
	@${BUILD_DIR}/${BINARY_NAME}

## docker-build: Build Docker image for autobox-engine
docker-build:
	@echo "Building autobox-engine Docker image..."
	@cd ../autobox-engine && docker build -t autobox-engine:latest .
	@echo "Docker image built: autobox-engine:latest"

## reinstall: Clean, build and install in one command
reinstall: clean build install
	@echo "Reinstall complete!"

## release: Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	@mkdir -p ${BUILD_DIR}/release
	
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/release/${BINARY_NAME}-linux-amd64 ${MAIN_PATH}
	
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/release/${BINARY_NAME}-linux-arm64 ${MAIN_PATH}
	
	@echo "Building for Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/release/${BINARY_NAME}-darwin-amd64 ${MAIN_PATH}
	
	@echo "Building for Darwin ARM64..."
	@GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/release/${BINARY_NAME}-darwin-arm64 ${MAIN_PATH}
	
	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/release/${BINARY_NAME}-windows-amd64.exe ${MAIN_PATH}
	
	@echo "Release binaries built in ${BUILD_DIR}/release/"