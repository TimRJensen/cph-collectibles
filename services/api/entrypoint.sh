#!/bin/sh
set -eu

ASSET_DIR="${ASSET_DIR:-/app/assets}"
DATA_DIR="${DATA_DIR:-/app/assets/data}"

mkdir -p "$ASSET_DIR"
mkdir -p "$DATA_DIR"
# chown -R app:app "$ASSET_DIR"

if [ ! -f "$DATA_DIR/posters.csv" ]; then
    wget -O "$DATA_DIR/posters.csv" \
        https://github.com/TimRJensen/cph-collectibles/releases/download/v1/posters.csv
fi

if [ ! -f "$DATA_DIR/posters.zip" ]; then
    wget -O "$DATA_DIR/posters.zip" \
        https://github.com/TimRJensen/cph-collectibles/releases/download/v1/posters.zip
fi

chown -R app:app "$ASSET_DIR"

CMD="${1:-}"
if [ -z "$CMD" ]; then
    echo "No command provided" >&2
    exit 1
fi
shift || true

exec su app -s /bin/sh -c 'exec "$@"' sh "$CMD" "$@"