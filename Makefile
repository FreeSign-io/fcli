.PHONY: build test lint vet tidy install run clean

GO         ?= go
BIN        := fcli
PKG        := github.com/FreeSign-io/fcli
VERSION    ?= dev
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE       := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
              -X $(PKG)/internal/version.Version=$(VERSION) \
              -X $(PKG)/internal/version.Commit=$(COMMIT) \
              -X $(PKG)/internal/version.Date=$(DATE)

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) .

test:
	$(GO) test ./... -race -count=1

lint:
	golangci-lint run

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install -ldflags "$(LDFLAGS)" .

run: build
	./$(BIN)

clean:
	rm -f $(BIN) $(BIN)-* dist/* 2>/dev/null || true
