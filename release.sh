#!/bin/bash
set -e

readonly BIN="micro-sql"

NOW=$(date +%Y%m%d.%H%M%S)
HASH=$(git log -1 --pretty=%H | cut -c-10)
HOSTNAME=$(hostname -s)
echo "${NOW}:${HASH}:${HOSTNAME}" > releases/.build

echo "Building binaries ($(cat releases/.build))"
echo "  Linux.."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.build=${HASH}" -o releases/${BIN}.linux ./cmd/main.go
echo "  Darwin.."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.build=${HASH}" -o releases/${BIN}.darwin ./cmd/main.go
echo "  Windows.."
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.build=${HASH}" -o releases/${BIN}.exe ./cmd/main.go

echo "Release binaries available in 'releases/'"
