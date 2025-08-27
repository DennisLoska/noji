BINARY := noji

build:
	GO111MODULE=on go build -ldflags "-X github.com/dennis/noji/internal/version.Version=$(VERSION) -X github.com/dennis/noji/internal/version.Commit=$(COMMIT) -X github.com/dennis/noji/internal/version.Date=$(DATE) -s -w" -o bin/$(BINARY) ./cmd/noji

run:
	GO111MODULE=on go run ./cmd/noji --help

clean:
	rm -rf bin
