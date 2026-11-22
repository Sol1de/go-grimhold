BIN := bin/grimhold

ifeq ($(OS),Windows_NT)
SHELL := cmd.exe
BIN := $(BIN).exe
RM_BIN := if exist bin rmdir /s /q bin
else
RM_BIN := rm -rf bin
endif

.PHONY: build run fmt lint vet test clean

build:
	go build -o $(BIN) ./cmd/grimhold

run:
	go run ./cmd/grimhold

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	$(RM_BIN)
