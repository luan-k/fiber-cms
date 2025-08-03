#!/bin/sh

set -e

echo "Starting the application..."
echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "Starting the server..."
exec "$@"