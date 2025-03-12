#!/bin/bash
set -e  # Exit on error

BIN="micro-mysql"

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

echo "Setup complete. Run './${BIN} -u demo -p demopasswd -h localhost -P 3306 schema' to start."
