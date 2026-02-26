.PHONY: build-tui build-gui build-all dev-tui dev-gui tidy test

build-tui:
	cd tor-tui && go build -o bin/tor-tui.exe .

build-gui:
	cd tor-gui && wails build -clean

build-all: build-tui build-gui

dev-tui:
	cd tor-tui && go run .

dev-gui:
	cd tor-gui && wails dev

tidy:
	cd core    && go mod tidy
	cd tor-tui && go mod tidy
	cd tor-gui && go mod tidy

test:
	cd core && go test ./...