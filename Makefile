# Default operation
.DEFAULT_GOAL := build

# Required Variables
VERSION := $(shell cat .version)
GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
#GOARCH ?= $(shell uname -m) # Does not work in Docker Alpine
GOARCH ?= arm64
HASH := $(shell cat .hash 2>/dev/null || echo "unknown")
OUTPUT := micro-sql
AUTHOR := ronaldbradford

# Dynamic Build Flags
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.build=$(HASH)'

# Targets
.PHONY: hash build build-docker setup test clean

hash:
	@echo "Generating .hash file..."
	@git log -1 --pretty=%h > .hash

setup:
	@if [ ! -f "go.mod" ]; then \
		echo "Initializing Go module..."; \
		go mod init "github.com/ronaldbradford/$(OUTPUT)"; \
	fi
	@echo "Installing dependencies..."
	go mod tidy

build: setup
	@echo "Building for $(GOOS)/$(GOARCH) v$(VERSION)-$(HASH)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o bin/$(OUTPUT) ./cmd
	@echo "Creating symbolic links..."
	ln -sf $(OUTPUT) bin/micro-mysql
	ln -sf $(OUTPUT) bin/micro-psql

build-docker:
	docker build --tag $(AUTHOR)/$(OUTPUT):latest --tag $(AUTHOR)/$(OUTPUT):$(VERSION) .

test:
	@echo "Running tests in cmd/..."
	go clean -testcache
	@go test -v ./cmd/...

clean:
	@echo "Cleaning build artifacts..."
	#go clean -cache -modcache
	rm -rf bin/
