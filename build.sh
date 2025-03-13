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
go build -o "${BASE_DIR}/bin/${BIN}" "${BASE_DIR}/cmd"

# Create symlinks for MySQL and PostgreSQL modes
ln -sf bin/${BIN} "${BASE_DIR}/bin/micro-mysql"
ln -sf bin/${BIN} "${BASE_DIR}/bin/micro-psql"

echo "Build complete. Use ./bin/micro-mysql or ./bin/micro-psql."
