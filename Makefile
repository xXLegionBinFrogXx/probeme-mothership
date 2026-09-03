GO ?= go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X github.com/xXLegionBinFrogXx/probeme-mothership/internal/buildinfo.Version=$(VERSION) \
	-X github.com/xXLegionBinFrogXx/probeme-mothership/internal/buildinfo.Commit=$(COMMIT) \
	-X github.com/xXLegionBinFrogXx/probeme-mothership/internal/buildinfo.Date=$(DATE)

BINARY := bin/probeme-mothership

PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
UNITDIR := $(PREFIX)/lib/systemd/system

.PHONY: all build fakeprovider test lint integration run install uninstall release clean

all: build

# cgo is required to talk to libprobeme; every target below runs with CGO_ENABLED=1.
build: fakeprovider
	@pkg-config --cflags --libs probeme >/dev/null 2>&1 || { \
		echo "error: pkg-config probeme failed; install libprobeme (cmake --install) and set PKG_CONFIG_PATH" >&2; \
		exit 1; \
	}
	CGO_ENABLED=1 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/probeme-mothership

fakeprovider:
	./test/fakeprovider/build.sh

test: fakeprovider
	CGO_ENABLED=0 $(GO) build ./...
	CGO_ENABLED=0 $(GO) test ./internal/metrics/
	CGO_ENABLED=1 $(GO) test -race ./...

lint:
	golangci-lint run

integration: build
	CGO_ENABLED=1 $(GO) test -race -tags integration ./test/integration/ -v

run: build
	./$(BINARY) --providers=$(PROVIDERS) --log.level=debug

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(BINARY) $(DESTDIR)$(BINDIR)/
	install -d $(DESTDIR)$(UNITDIR)
	install -m 0644 packaging/probeme-mothership.service $(DESTDIR)$(UNITDIR)/
	@echo "installed $(DESTDIR)$(BINDIR)/probeme-mothership"
	@echo "installed $(DESTDIR)$(UNITDIR)/probeme-mothership.service"

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/probeme-mothership
	rm -f $(DESTDIR)$(UNITDIR)/probeme-mothership.service

release:
	./scripts/release.sh $(VERSION)

clean:
	rm -rf bin test/fakeprovider/build release
