#!/bin/bash
set -e  # Exit on error

BIN="micro-sql"

# Initialize Go module if not exists
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init "${BIN}"
fi

# Install dependencies
echo "Installing dependencies..."
go get -u github.com/go-sql-driver/mysql
go mod tidy

# Build the application
echo "Building the application..."
go build -o "${BIN}"
#go build -o micro-sql main.go

# Create symlinks for MySQL and PostgreSQL modes
ln -sf ${BIN} micro-mysql
ln -sf ${BIN} micro-psql

echo "Build complete. Use ./micro-mysql or ./micro-psql."
