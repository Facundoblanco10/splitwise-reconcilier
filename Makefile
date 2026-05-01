BINARY := reconciler
MONTH  ?= $(shell date +%Y-%m)

.PHONY: run build clean help

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
