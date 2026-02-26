APP     := tor-tui
VERSION := 0.1.0
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: build run install clean tidy deps

## build: Compile tor-tui binary
build:
	go build $(LDFLAGS) -o bin/$(APP) .

## run: Run tor-tui directly (requires Tor running on default ports)
run:
	go run . 

## install: Install to ~/go/bin
install:
	go install $(LDFLAGS) .

## tidy: Update go.mod and go.sum
tidy:
	go mod tidy

## deps: Download dependencies
deps:
	go mod download

## clean: Remove build artifacts
clean:
	rm -rf bin/

## cross: Build for Linux, macOS, Windows
cross: clean
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP)-linux-amd64   .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP)-macos-amd64   .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP)-macos-arm64   .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP)-windows.exe   .

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
