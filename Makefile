# castctl build/install (macOS/Linux; use install.ps1 on Windows).
VERSION ?= 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)
PREFIX  ?= $(HOME)/.local/bin

.PHONY: build install test clean dist

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o dist/castctl ./

# Installs to PREFIX (default ~/.local/bin). Ensure PREFIX is on your PATH.
install: build
	mkdir -p $(PREFIX)
	cp dist/castctl $(PREFIX)/castctl
	@echo "Installed to $(PREFIX)/castctl"

test:
	go test ./...

# Cross-compile every release target into dist/.
dist:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/castctl-windows-amd64.exe ./
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/castctl-darwin-arm64 ./
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/castctl-darwin-amd64 ./
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/castctl-linux-amd64 ./

clean:
	rm -rf dist
