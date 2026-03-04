BINARY := meta-cli
PKG    := github.com/ygncode/meta-cli/cmd/meta

.PHONY: build install test lint tidy clean

build:
	go build -o $(BINARY) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

lint:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
