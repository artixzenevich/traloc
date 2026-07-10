BINARY=bin/traloc

.PHONY: build test lint clean run release snapshot

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/traloc

test:
	go test -v -race ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/ dist/

run:
	go run ./cmd/traloc $(ARGS)

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean
