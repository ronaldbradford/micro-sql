# Default operation
.DEFAULT_GOAL := build

# Required Variables
VERSION := $(shell cat .version)
GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH ?= $(shell uname -m)
HASH := $(shell git log -1 --pretty=%h)
OUTPUT := micro-sql

# Dynamic Build Flags
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.build=$(HASH)'

# Targets
.PHONY: build setup clean

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

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
