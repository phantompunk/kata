.PHONY: all build clean test

binary = kata

all:
build:
	go build -o=$(GOPATH)/bin/$(binary)

test:
	go test -v -race -buildvcs ./...

test/cover:
	go test -v -race -buildvcs -cover -coverprofile c.out
	go tool cover -html=c.out -o coverage.html

clean:
	rm -rf $(GOPATH)/bin/$(binary) c.out coverage.html

