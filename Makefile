# Makefile for Veda IO (Launcher)

VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

.PHONY: all build build-engine build-ui clean

all: build

build: build-engine build-ui
	@echo "Building Veda IO Launcher..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o vedaio.exe .
	upx --best --lzma vedaio.exe

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
	rm -f vedaio.exe
	$(MAKE) -C ../veda-engine clean
	$(MAKE) -C ../veda-ui clean
