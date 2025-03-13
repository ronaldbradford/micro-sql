#!/bin/bash
set -e

readonly BIN="micro-sql"

echo "Building Linux binary"
GOOS=linux GOARCH=amd64 go build -o releases/linux/${BIN} ./cmd/main.go
echo "Building Darwin binary"
GOOS=darwin GOARCH=amd64 go build -o releases/darwin/${BIN} ./cmd/main.go
echo "Building Windows binary"
GOOS=windows GOARCH=amd64 go build -o releases/windows/${BIN}.exe ./cmd/main.go

echo "Release binaries available in 'releases/'"
