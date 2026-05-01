BINARY := reconciler
MONTH  ?= $(shell date +%Y-%m)

.PHONY: run build clean setup test help

## setup: interactive first-time configuration (.env and config.yaml)
setup:
	@bash setup.sh

## test: run all unit tests
test:
	go test ./...

## test-v: run all unit tests with verbose output
test-v:
	go test -v ./...

## run: fetch and generate report for MONTH (default: current month)
run:
	go run ./cmd/reconciler --month $(MONTH)

## build: compile the binary
build:
	go build -o $(BINARY) ./cmd/reconciler

## clean: remove the compiled binary and generated reports
clean:
	rm -f $(BINARY)
	rm -f output/*.xlsx

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
