#!/bin/sh
set -e

echo "Populating database..."
./populate

echo "Starting server..."
exec ./server
