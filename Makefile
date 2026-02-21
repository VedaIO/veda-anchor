# Makefile for Veda Anchor Anchor(Launcher)

VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build generate build-engine build-ui clean fmt

all: build

generate:
	@echo "Generating version info..."
	go generate

build: generate build-engine build-ui
	@echo "Building Veda Anchor AnchorLauncher..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -H=windowsgui -X main.Version=$(VERSION)" -o veda-anchor-anchor-anchor.exe .
	upx --best --lzma veda-anchor-anchor-anchor.exe

build-engine:
	@echo "Building Veda Anchor AnchorEngine..."
	$(MAKE) -C ../veda-anchor-anchor-anchor-engine build
	@mkdir -p bin
	cp ../veda-anchor-anchor-anchor-engine/bin/veda-anchor-anchor-anchor-engine.exe bin/veda-anchor-anchor-anchor-engine.exe

build-ui:
	@echo "Building Veda Anchor AnchorUI..."
	$(MAKE) -C ../veda-anchor-anchor-anchor-ui build
	@mkdir -p bin
	cp ../veda-anchor-anchor-anchor-ui/build/bin/veda-anchor-anchor-anchor-ui.exe bin/veda-anchor-anchor-anchor-ui.exe

fmt:
	@echo "Formatting code..."
	go fmt ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f veda-anchor-anchor-anchor.exe
	rm -f resource.syso
	$(MAKE) -C ../veda-anchor-anchor-anchor-engine clean
	$(MAKE) -C ../veda-anchor-anchor-anchor-ui clean
