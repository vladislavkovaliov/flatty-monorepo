#!/bin/sh
set -e

# These migrations depend on the better-auth tables ("user"), which are created
# by nest-graphql at startup. The compose service waits for its healthcheck.
export PGPASSWORD="$POSTGRES_PASSWORD"

psql_db() {
  psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    "$@"
}

echo "Waiting for better-auth schema..."
until [ "$(psql_db -tAc "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'user'")" = "1" ]; do
  sleep 1
done
echo "better-auth schema is ready"

if [ "$(psql_db -tAc "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'user_settings'")" = "1" ]; then
  echo "user_settings already exists, skipping 004_add_user_settings.sql"
else
  echo "Applying 004_add_user_settings.sql..."
  psql_db -v ON_ERROR_STOP=1 -f /migrations/004_add_user_settings.sql
fi

echo "Applying 007_add_user_id_to_expenses.sql..."
psql_db -v ON_ERROR_STOP=1 -f /migrations/007_add_user_id_to_expenses.sql

echo "Auth migrations complete"
