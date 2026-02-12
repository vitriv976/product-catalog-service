#!/bin/bash

# Setup script for initializing Spanner database
# This script helps set up the database for local development

set -e

PROJECT="test-project"
INSTANCE="test-instance"
DATABASE="test-database"
EMULATOR_HOST="localhost:9010"

echo "=== Product Catalog Service - Database Setup ==="
echo ""
echo "Configuration:"
echo "  Project: $PROJECT"
echo "  Instance: $INSTANCE"
echo "  Database: $DATABASE"
echo "  Emulator: $EMULATOR_HOST"
echo ""

# Set emulator host
export SPANNER_EMULATOR_HOST=$EMULATOR_HOST

# Check if spanner-cli is installed
if ! command -v spanner-cli &> /dev/null; then
    echo "ERROR: spanner-cli not found!"
    echo ""
    echo "To install spanner-cli, run:"
    echo "  go install github.com/cloudspannerecosystem/spanner-cli@latest"
    echo ""
    exit 1
fi

echo "✓ spanner-cli found"
echo ""

# Create database and run migrations
echo "Creating database and running migrations..."
echo ""

spanner-cli -p $PROJECT -i $INSTANCE -d $DATABASE < migrations/001_initial_schema.sql

echo ""
echo "✓ Database created and migrations applied successfully!"
echo ""
echo "You can now run the server with:"
echo "  make run"
echo ""
echo "Or run tests with:"
echo "  make test"
echo ""
