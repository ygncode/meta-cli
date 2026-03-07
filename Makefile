BINARY := meta-cli
PKG    := github.com/ygncode/meta-cli/cmd/meta

.PHONY: build install test lint tidy clean install-skill

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

install-skill:
	mkdir -p ~/.openclaw/workspace/skills/
	cp -r skill/meta-cli-fb ~/.openclaw/workspace/skills/
