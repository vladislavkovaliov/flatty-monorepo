#!/bin/sh
set -e

echo "Verifying database tables..."

TABLES=$(PGPASSWORD="$POSTGRES_PASSWORD" psql \
  -h "$POSTGRES_HOST" \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  -t -c "SELECT table_name FROM information_schema.tables WHERE table_schema='public' ORDER BY table_name;")

echo "Tables found:"
echo "$TABLES"

EXPECTED="expenses
categories
resident_locations"

for table in $(echo "$EXPECTED"); do
  if echo "$TABLES" | grep -qi "$table"; then
    COUNT=$(PGPASSWORD="$POSTGRES_PASSWORD" psql \
      -h "$POSTGRES_HOST" \
      -U "$POSTGRES_USER" \
      -d "$POSTGRES_DB" \
      -t -c "SELECT COUNT(*) FROM $table;")
    echo "  $table: $COUNT rows"
  else
    echo "  ERROR: $table table not found!"
    exit 1
  fi
done

echo "Database verification passed"
