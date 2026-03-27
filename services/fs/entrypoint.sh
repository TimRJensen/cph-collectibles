#!/bin/sh
set -eu

ASSET_DIR="${ASSET_DIR:-/app/assets}"

mkdir -p "$ASSET_DIR"
chown -R app:app "$ASSET_DIR"

touch "$ASSET_DIR/.write-test"
rm -f "$ASSET_DIR/.write-test"

CMD="$1"
shift || true

exec su app -s /bin/sh -c 'exec "$@"' sh "$CMD" "$@"