#!/usr/bin/env bash

set -euo pipefail

# Minimal launcher for running the Flask app under Gunicorn (WSGI).
# Usage:
#   APP_HOST=0.0.0.0 APP_PORT=8080 APP_WORKERS=4 ./start.sh

HOST="${APP_HOST:-0.0.0.0}"
PORT="${APP_PORT:-8080}"
# recommendation: use 2 workers per CPU
WORKERS="${APP_WORKERS:-1}"
THREADS="${APP_THREADS:-2}"
BACKLOG="${APP_BACKLOG:-2048}"

printf "Starting gunicorn with\n - Bind: %s:%s\n - Workers: %s\n - Threads: %s\n - Backlog: %s\n\n" "${HOST}" "${PORT}" "${WORKERS}" "${THREADS}" "${BACKLOG}"

exec gunicorn \
  --workers "${WORKERS}" \
  --bind "${HOST}:${PORT}" \
  --threads "${THREADS}" \
  --backlog "${BACKLOG}" \
  "internal.webserver:create_app()" \
