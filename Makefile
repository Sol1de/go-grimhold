BIN := bin/grimhold

.PHONY: build fmt lint vet test clean

build:
	mkdir -p bin
	go build -o $(BIN) ./cmd/grimhold

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -rf bin
