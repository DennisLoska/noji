APP := noji
MAIN := ./cmd/noji

VERSION ?= $(shell git describe --tags --always --dirty=never 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
  -X github.com/dennisloska/noji/internal/version.Version=$(VERSION) \
  -X github.com/dennisloska/noji/internal/version.Commit=$(COMMIT) \
  -X github.com/dennisloska/noji/internal/version.Date=$(DATE)

.PHONY: build install release-install run clean

build:
	GO111MODULE=on CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o bin/$(APP) $(MAIN)

install:
	GO111MODULE=on CGO_ENABLED=0 go install -ldflags '$(LDFLAGS)' $(MAIN)

# Force an exact version string, e.g., make release-install VERSION=v0.1.0
release-install:
	@[ -n "$(VERSION)" ] || (echo "VERSION is required"; exit 1)
	GO111MODULE=on CGO_ENABLED=0 go install -ldflags '-s -w \
		-X github.com/dennisloska/noji/internal/version.Version=$(VERSION) \
		-X github.com/dennisloska/noji/internal/version.Commit=$(COMMIT) \
		-X github.com/dennisloska/noji/internal/version.Date=$(DATE)' \
		$(MAIN)

run:
	GO111MODULE=on go run ./cmd/noji --help

clean:
	rm -rf bin

