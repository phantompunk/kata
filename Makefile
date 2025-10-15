.PHONY: help build clean test install dev install-dev uninstall-dev

BINARY = kata
AIR_BINARY = ./tmp/kata
INSTALL_PATH=$(HOME)/.local/bin

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
	go build -o=$(GOPATH)/bin/$(BINARY)

## install: Install the kata binary
install:
	go install ./cmd . 

dev: install-dev
	@air

## install-dev: Install the kata development binary
install-dev:
	@mkdir -p $(INSTALL_PATH)
	@ln -sf $(PWD)/$(AIR_BINARY) $(INSTALL_PATH)/$(BINARY)
	@echo "Symlinked $(INSTALL_PATH)/$(BINARY) -> $(PWD)/$(AIR_BINARY)"

## uninstall-dev: Remove the kata development binary
uninstall-dev:
	@rm -f $(INSTALL_PATH)/$(BINARY)
	@echo "Removed $(INSTALL_PATH)/$(BINARY)"

## test: Run tests
test:
	go test -v -race ./...

## test/cover: Run tests with coverage and generate HTML report
test/cover:
	go test -v -race -buildvcs -cover -coverprofile c.out
	go tool cover -html=c.out -o coverage.html

## clean: Remove the binary and coverage files
clean:
	rm -rf $(GOPATH)/bin/$(BINARY) c.out coverage.html

