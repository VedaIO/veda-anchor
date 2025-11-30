# Makefile for procguard

# Use git describe to get a version string.
# Example: v1.0.0-3-g1234567
# Fallback to 'dev' if not in a git repository.
VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build dev fmt clean

all: build

build:
	@echo "Building ProcGuard for windows..."
	cd wails-app && wails build -platform windows/amd64 -ldflags="-X main.version=$(VERSION)"

build-debug:
	@echo "Building ProcGuard for windows..."
	cd wails-app && wails build -platform windows/amd64 -debug

fmt:
	@echo "Formatting code..."
	cd wails-app && go fmt ./...
	cd wails-app/frontend && npm run format

lint:
	cd wails-app && GOOS=windows golangci-lint run
	cd wails-app/frontend && npm run lint

clean:
	@echo "Cleaning..."
	rm -rf wails-app/build/bin
	rm -rf wails-app/frontend/dist
	rm -rf wails-app/frontend/wailsjs
