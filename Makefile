# Makefile for procguard

# Use git describe to get a version string.
# Example: v1.0.0-3-g1234567
# Fallback to 'dev' if not in a git repository.
VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build dev fmt clean

all: build

build:
	@echo "Building ProcGuard for windows..."
	export PATH=$(shell $(GO_CMD) env GOPATH)/bin:$$PATH && cd procguard-wails && wails build -platform windows/amd64 -ldflags="-X main.version=$(VERSION)"

fmt:
	@echo "Formatting code..."
	cd procguard-wails && go fmt ./...
	cd procguard-wails/frontend && npm run format

clean:
	@echo "Cleaning..."
	rm -rf procguard-wails/build/bin
