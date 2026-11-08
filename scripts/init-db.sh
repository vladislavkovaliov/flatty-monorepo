#!/bin/sh
set -e

echo "Waiting for PostgreSQL..."
until pg_isready -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" 2>/dev/null; do
  sleep 1
done
echo "PostgreSQL is ready"

echo "Running base schema..."
PGPASSWORD="$POSTGRES_PASSWORD" psql \
  -h "$POSTGRES_HOST" \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  -f /db.sql

# The applications registry is not part of db.sql; it lives in go-api
# migrations. Run the idempotent ones that do not depend on better-auth tables.
echo "Applying application registry migrations..."
for migration in \
  /migrations/006_add_applications.sql \
  /migrations/008_add_vue_torrent_bot_application.sql; do
  echo "  -> $migration"
  PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -v ON_ERROR_STOP=1 \
    -f "$migration"
done

echo "Ensuring torrent_bot database (apps/python-torrent-bot)..."
if [ "$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "SELECT 1 FROM pg_database WHERE datname='torrent_bot'")" != "1" ]; then
  PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "CREATE DATABASE torrent_bot"
fi

echo "Migration complete"
