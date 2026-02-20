# Makefile for Veda IO (Launcher)

VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build generate build-engine build-ui clean

all: build

generate:
	@echo "Generating version info..."
	go generate

build: generate build-engine build-ui
	@echo "Building Veda Launcher..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -H=windowsgui -X main.Version=$(VERSION)" -o veda.exe .
	upx --best --lzma veda.exe

build-engine:
	@echo "Building Veda Engine..."
	$(MAKE) -C ../veda-engine build
	@mkdir -p bin
	cp ../veda-engine/bin/veda-engine.exe bin/veda-engine.exe

build-ui:
	@echo "Building Veda UI..."
	$(MAKE) -C ../veda-ui build
	@mkdir -p bin
	cp ../veda-ui/build/bin/veda-ui.exe bin/veda-ui.exe

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f veda.exe
	rm -f resource.syso
	$(MAKE) -C ../veda-engine clean
	$(MAKE) -C ../veda-ui clean
