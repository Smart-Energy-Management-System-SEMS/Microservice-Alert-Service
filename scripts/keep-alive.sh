#!/usr/bin/env bash
set -euo pipefail

TARGET_URL="${TARGET_URL:-}"
PING_PATH="${PING_PATH:-/api/v1/alerts}"
INTERVAL_SECONDS="${INTERVAL_SECONDS:-600}"

if [[ -z "$TARGET_URL" ]]; then
  echo "Error: define TARGET_URL, por ejemplo https://tu-servicio.onrender.com"
  exit 1
fi

TARGET_URL="${TARGET_URL%/}"
URL="${TARGET_URL}${PING_PATH}"

echo "Keep-alive iniciado para: ${URL}"
echo "Intervalo: ${INTERVAL_SECONDS}s"

while true; do
  if curl -fsS --max-time 20 "$URL" > /dev/null; then
    echo "$(date -u +'%Y-%m-%dT%H:%M:%SZ') ping ok"
  else
    echo "$(date -u +'%Y-%m-%dT%H:%M:%SZ') ping error"
  fi
  sleep "$INTERVAL_SECONDS"
done
