#!/bin/bash
set -e  # Exit on error

readonly BIN="micro-sql"

BASE_DIR=$(dirname "$0")
mkdir -p "${BASE_DIR}/bin"

# Initialize Go module if not exists
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init "github.com/ronaldbradford/${BIN}"
fi

echo "Installing dependencies..."
go mod tidy

# Build the application
echo "Building the application..."
HASH=$(git log -1 --pretty=%H | cut -c-10)
go build -ldflags "-X main.build=${HASH}" -o "${BASE_DIR}/bin/${BIN}" "${BASE_DIR}/cmd"

# Create symlinks for MySQL and PostgreSQL modes
ln -sf ${BIN} "${BASE_DIR}/bin/micro-mysql"
ln -sf ${BIN} "${BASE_DIR}/bin/micro-psql"

echo "Build complete. Use ./bin/micro-mysql or ./bin/micro-psql."
