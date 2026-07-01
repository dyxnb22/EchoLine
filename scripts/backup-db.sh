#!/usr/bin/env bash
# Backup EchoLine PostgreSQL database to a timestamped SQL file.
set -euo pipefail

DATABASE_URL="${DATABASE_URL:-postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable}"
OUT_DIR="${1:-./backups}"
mkdir -p "$OUT_DIR"
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
OUT_FILE="$OUT_DIR/echoline-$STAMP.sql"

echo "Backing up to $OUT_FILE"
pg_dump "$DATABASE_URL" > "$OUT_FILE"
echo "Done: $(wc -c < "$OUT_FILE") bytes"
