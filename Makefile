.PHONY: help build clean test install

binary = kata

# Set the default goal
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "Usage:\n  make [command]"
	@echo ""
	@echo "Commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## build: Build the kata binary
build:
	go build -o=$(GOPATH)/bin/$(binary)

## install: Install the kata binary
install:
	go install ./cmd/kata

## test: Run tests
test:
	go test -v -race ./...

## test/cover: Run tests with coverage and generate HTML report
test/cover:
	go test -v -race -buildvcs -cover -coverprofile c.out
	go tool cover -html=c.out -o coverage.html

## clean: Remove the binary and coverage files
clean:
	rm -rf $(GOPATH)/bin/$(binary) c.out coverage.html

