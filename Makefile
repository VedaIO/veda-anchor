# Makefile for Veda Anchor (Launcher)

VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build generate build-engine build-agent build-ui clean fmt

all: build

init:
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

generate:
	@echo "Generating version info..."
	go generate

build: generate build-engine build-agent build-ui
	@echo "Building Veda Anchor Launcher..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -H=windowsgui -X main.Version=$(VERSION)" -o veda-anchor.exe .
	upx --best --lzma veda-anchor.exe

build-engine:
	@echo "Building Veda Anchor Engine..."
	$(MAKE) -C ../veda-anchor-engine build
	@mkdir -p bin
	cp ../veda-anchor-engine/bin/veda-anchor-engine.exe bin/veda-anchor-engine.exe

build-agent:
	@echo "Building Veda Anchor Agent..."
	$(MAKE) -C ../veda-anchor-agent build
	@mkdir -p bin
	cp ../veda-anchor-agent/bin/veda-anchor-agent.exe bin/veda-anchor-agent.exe

build-ui:
	@echo "Building Veda Anchor UI..."
	$(MAKE) -C ../veda-anchor-ui build
	@mkdir -p bin
	cp ../veda-anchor-ui/build/bin/veda-anchor-ui.exe bin/veda-anchor-ui.exe

fmt:
	@echo "Formatting code..."
	go fmt ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f veda-anchor.exe
	rm -f resource.syso
	$(MAKE) -C ../veda-anchor-engine clean
	$(MAKE) -C ../veda-anchor-agent clean
	$(MAKE) -C ../veda-anchor-ui clean

lint:
	CGO_ENABLED=0 GOOS=windows golangci-lint run
