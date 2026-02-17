# Makefile for wireguard-tui

BINARY_NAME=wireguard-tui
BUILD_DIR=dist
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "v0.1.0")
LDFLAGS=-ldflags "-X main.version=${VERSION} -s -w"

.PHONY: all build build-linux clean package

all: clean build

build:
	go build ${LDFLAGS} -o ${BINARY_NAME} ./cmd/wireguard-tui/main.go

# Cross-compilation for requested Linux platforms
build-linux:
	@mkdir -p ${BUILD_DIR}
	# amd64
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/wireguard-tui/main.go
	# arm64
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-arm64 ./cmd/wireguard-tui/main.go

clean:
	rm -rf ${BINARY_NAME} ${BUILD_DIR}

# Help target
help:
	@echo "Available targets:"
	@echo "  build         Build the binary for the current OS/Arch"
	@echo "  build-linux   Cross-compile for linux/amd64 and linux/arm64"
	@echo "  clean         Remove build artifacts"
	@echo "  package       Instructions for packaging with GoReleaser"

package:
	@echo "To generate .deb and Arch packages, run:"
	@echo "  goreleaser release --snapshot --clean"
