#!/bin/sh
set -e

echo "Waiting for PostgreSQL..."
until pg_isready -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" 2>/dev/null; do
  sleep 1
done
echo "PostgreSQL is ready"

echo "Running migrations..."
PGPASSWORD="$POSTGRES_PASSWORD" psql \
  -h "$POSTGRES_HOST" \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  -f /db.sql

echo "Ensuring torrent_bot database (apps/python-torrent-bot)..."
if [ "$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "SELECT 1 FROM pg_database WHERE datname='torrent_bot'")" != "1" ]; then
  PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "CREATE DATABASE torrent_bot"
fi

echo "Migration complete"
