#!/usr/bin/env bash
set -e

echo "Starting postgres via original entrypoint..."
/usr/local/bin/docker-entrypoint.sh "$@" &

PID=$!

echo "Waiting for postgres..."
until pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; do
  sleep 1
done

if [ ! -f "$PGDATA/.schema_done" ]; then
  echo "Running init SQL..."
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f /init/schema.sql
  touch "$PGDATA/.schema_done"
fi

echo "Handing over to postgres..."
wait "$PID"