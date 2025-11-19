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

# Prometheus multiprocess setup
export PROMETHEUS_MULTIPROC_DIR="${PROMETHEUS_MULTIPROC_DIR:-/tmp/prometheus}"
mkdir -p "${PROMETHEUS_MULTIPROC_DIR}"
# Clean up old metrics files before starting workers
find "${PROMETHEUS_MULTIPROC_DIR}" -type f -name '*.db' -delete 2>/dev/null || true
find "${PROMETHEUS_MULTIPROC_DIR}" -type f -name '*.pid' -delete 2>/dev/null || true

printf "Starting gunicorn with\n - Bind: %s:%s\n - Workers: %s\n - Threads: %s\n - Backlog: %s\n\n" "${HOST}" "${PORT}" "${WORKERS}" "${THREADS}" "${BACKLOG}"

exec gunicorn \
  --workers "${WORKERS}" \
  --bind "${HOST}:${PORT}" \
  --threads "${THREADS}" \
  --backlog "${BACKLOG}" \
  --config gunicorn_conf.py \
  "internal.webserver:create_app()" \
