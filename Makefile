BINARY := noji

build:
	GO111MODULE=on go build -o bin/$(BINARY) ./cmd/noji

run:
	GO111MODULE=on go run ./cmd/noji --help

clean:
	rm -rf bin
