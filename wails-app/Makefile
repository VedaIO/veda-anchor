# Makefile for procguard

# Use git describe to get a version string.
# Example: v1.0.0-3-g1234567
# Fallback to 'dev' if not in a git repository.
VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build dev fmt clean

all: build
# DO NOT ADD -s FLAG OR ELSE ANTI VIRUS WILL TRIGGER
build:
	@echo "Building ProcGuard for windows..."
	cd frontend && bun run build
	wails generate module
	CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu -Wl,--subsystem,windows" GOOS=windows GOARCH=amd64 go build -tags desktop,production -ldflags="-w -X main.version=$(VERSION) -H windowsgui" -o build/bin/ProcGuard.exe

build-debug:
	@echo "Building ProcGuard for windows (debug)..."
	cd frontend && bun run build
	wails generate module
	CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu" GOOS=windows GOARCH=amd64 go build -tags desktop,production -ldflags="-X main.version=$(VERSION)" -o build/bin/ProcGuard.exe

fmt:
	@echo "Formatting code..."
	go fmt ./...
	cd frontend && bun run format

lint:
	GOOS=windows golangci-lint run
	cd frontend && bun run lint

clean:
	@echo "Cleaning..."
	rm -rf build/bin
	rm -rf frontend/dist
	rm -rf frontend/wailsjs
