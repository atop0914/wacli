.PHONY: build clean install test

BINARY_NAME=wacli
DIST_DIR=./dist
BUILD_CMD=go build -tags sqlite_fts5 -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/wacli

build:
	@mkdir -p $(DIST_DIR)
	$(BUILD_CMD)

clean:
	rm -rf $(DIST_DIR)

install: build
	cp $(DIST_DIR)/$(BINARY_NAME) /usr/local/bin/

test:
	CGO_ENABLED=1 go test -tags sqlite_fts5 -v ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

.DEFAULT_GOAL := build
