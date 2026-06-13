.PHONY: help run build test clean

BINARY := bin/echodraw-server

help:
	@echo "EchoDraw Server"
	@echo "  make run       - run with go run"
	@echo "  make build     - build binary to $(BINARY)"
	@echo "  make test      - run unit tests"
	@echo "  make clean     - remove build artifacts"

run:
	go run ./cmd/server

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin
